package usecase

import (
	"fmt"
	"net/http"
	"time"
	"transaction-service/domain"

	"github.com/labstack/echo/v4"

	"github.com/dgrijalva/jwt-go"
)

type jwtUsecase struct {
	token domain.JwtToken
	redis domain.JwtTokenRepo
}

func NewJWTUseCase(token domain.JwtToken, redis domain.JwtTokenRepo) domain.JwtTokenUsecase {
	return &jwtUsecase{token: token, redis: redis}
}

func (j *jwtUsecase) GenerateToken(id int64, role, iin string) (string, error) {
	accessTokenClaims := jwt.MapClaims{}

	accessTokenClaims["id"] = id
	accessTokenClaims["role"] = role
	accessTokenClaims["iin"] = iin
	accessTokenClaims["iat"] = time.Now().Unix()
	accessTokenClaims["exp"] = time.Now().Add(j.token.AccessTtl).Unix()
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessTokenClaims)
	signedToken, err := accessToken.SignedString([]byte(j.token.AccessSecret))
	if err != nil {
		return "", &domain.LogError{"cannot create signed token", err, http.StatusInternalServerError}
	}
	return signedToken, nil
}

func (j *jwtUsecase) ParseTokenAndGetID(token string) (int64, error) {
	claims, err := j.ParseToken(token)
	if err != nil {
		return -1, &domain.LogError{"invalid token", err, http.StatusBadRequest}
	}
	id, ok := claims["id"].(float64)
	if !ok {
		return -1, &domain.LogError{"invalid token", fmt.Errorf("id not found from token"), http.StatusBadRequest}
	}
	return int64(id), nil
}

func (j *jwtUsecase) ParseTokenAndGetRole(token string) (string, error) {
	claims, err := j.ParseToken(token)
	if err != nil {
		return "", &domain.LogError{"invalid token", err, http.StatusBadRequest}
	}
	role, ok := claims["role"].(string)
	if !ok {
		return "", &domain.LogError{"invalid token", fmt.Errorf("role not found from token"), http.StatusBadRequest}
	}
	return role, nil
}

func (j *jwtUsecase) InsertToken(id int64, token string) error {

	key := fmt.Sprintf("user:%d", id)
	if err := j.redis.InsertTokenRepo(key, token, j.GetAccessTTL()); err != nil {
		return &domain.LogError{"cannot insert token", err, http.StatusInternalServerError}
	}
	return nil
}

func (j *jwtUsecase) FindToken(id int64, token string) (bool, error) {
	key := fmt.Sprintf("user:%d", id)

	ok, err := j.redis.FindTokenRepo(key, token)
	if err != nil {
		return false, &domain.LogError{"cannot find token", err, http.StatusBadRequest}
	}
	return ok, nil
}

func (j *jwtUsecase) GetAccessTTL() time.Duration {
	return j.token.AccessTtl
}

func (j *jwtUsecase) ParseToken(token string) (jwt.MapClaims, error) {
	JWTToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("failed to extract token metadata, unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(j.token.AccessSecret), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := JWTToken.Claims.(jwt.MapClaims)
	if ok && JWTToken.Valid {
		exp, ok := claims["exp"].(float64)
		if !ok {
			return nil, fmt.Errorf("field exp not found from token")
		}
		expiredTime := time.Unix(int64(exp), 0)
		if time.Now().After(expiredTime) {
			return nil, fmt.Errorf("token expired")
		}
	}
	return claims, nil
}

func (j *jwtUsecase) JWTErrorChecker(err error, c echo.Context) error {
	// Redirects to the signIn form.
	return c.Redirect(http.StatusMovedPermanently, c.Echo().Reverse("userSignInForm"))
}
