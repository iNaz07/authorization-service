package domain

import (
	"time"

	"github.com/go-redis/redis"
)

type JwtToken struct {
	AccessSecret string
	RedisConn    *redis.Client
	AccessTtl    time.Duration
}

type JwtTokenUsecase interface {
	GenerateToken(id int64, role string) (string, error)
	ParseTokenAndGetID(token string) (int64, error)
	ParseTokenAndGetRole(token string) (string, error)
	InsertToken(id int64, token string) error
	FindToken(id int64, token string) bool
	GetAccessTTL() time.Duration
}
