package usecase

import (
	"fmt"
	"time"
	"transaction-service/domain"

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
return 0, nil
}

func (j *jwtUsecase) InsertToken(id int64, token string) error {
	key := fmt.Sprintf("user:%d", id)
	return j.token.RedisConn.Set(key, token, 10*time.Minute).Err()
}

func (j *jwtUsecase) FindToken(id int64, token string) bool {
	key := fmt.Sprintf("user:%d", id)

	value, err := j.token.RedisConn.Get(key).Result()
	if err != nil {
		return false
	}
	return token == value
}
