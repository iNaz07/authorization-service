package domain

import (
	"time"

	"github.com/labstack/echo/v4"
)

type JwtToken struct {
	AccessSecret string
	// RedisConn    *redis.Client
	AccessTtl time.Duration
}

type JwtTokenUsecase interface {
	GenerateToken(id int64, role, iin string) (string, error)
	ParseTokenAndGetID(token string) (int64, error)
	ParseTokenAndGetRole(token string) (string, error)
	JWTErrorChecker(err error, c echo.Context) error
	GetAccessTTL() time.Duration
	InsertToken(id int64, token string) error
	FindToken(id int64, token string) (bool, error)
}

type JwtTokenRepo interface {
	InsertTokenRepo(key, token string, ttl time.Duration) error
	FindTokenRepo(key, token string) (bool, error)
}
