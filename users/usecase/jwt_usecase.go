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
}

func NewJWTUseCase(token domain.JwtToken) domain.JwtTokenUsecase {
	return &jwtUsecase{token: token}
}

func (j *jwtUsecase) GenerateToken(id int64, role string) (string, error) {
	accessTokenClaims := jwt.MapClaims{}

	accessTokenClaims["id"] = id
	accessTokenClaims["role"] = role
	accessTokenClaims["iat"] = time.Now().Unix()
	accessTokenClaims["exp"] = time.Now().Add(j.token.AccessTtl).Unix()
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessTokenClaims)
	signedToken, err := accessToken.SignedString([]byte(j.token.AccessSecret))
	return signedToken, err
}

func (j *jwtUsecase) ParseTokenAndGetID(token string) (int64, error) {
	claims, err := j.ParseToken(token)
	if err != nil {
		return -1, fmt.Errorf("invalid token: %w", err)
	}
	id, ok := claims["id"].(float64)
	if !ok {
		return -1, fmt.Errorf("id not found from token")
	}
	return int64(id), nil
}

func (j *jwtUsecase) ParseTokenAndGetRole(token string) (string, error) {
	claims, err := j.ParseToken(token)
	if err != nil {
		return "", fmt.Errorf("invalid token: %w", err)
	}
	role, ok := claims["role"].(string)
	if !ok {
		return "", fmt.Errorf("role not found from token")
	}
	return role, nil
}

func (j *jwtUsecase) InsertToken(id int64, token string) error {
	key := fmt.Sprintf("user:%d", id)
	return j.token.RedisConn.Set(key, token, 30*time.Minute).Err()
}

func (j *jwtUsecase) FindToken(id int64, token string) bool {
	key := fmt.Sprintf("user:%d", id)

	value, err := j.token.RedisConn.Get(key).Result()
	if err != nil {
		return false
	}
	return token == value
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
