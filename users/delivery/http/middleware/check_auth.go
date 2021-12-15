package middleware

import (
	"fmt"
	"transaction-service/domain"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Authorization struct {
	JwtUsecase domain.JwtTokenUsecase
}

func InitAuthorization(jwtuc domain.JwtTokenUsecase) *Authorization {
	return &Authorization{JwtUsecase: jwtuc}
}

func (a *Authorization) GetConfig() middleware.JWTConfig {
	return middleware.JWTConfig{
		TokenLookup:             "cookie:access-token",
		ParseTokenFunc:          a.CheckToken,
		ErrorHandlerWithContext: a.JwtUsecase.JWTErrorChecker,
	}
}

func (a *Authorization) CheckToken(auth string, c echo.Context) (interface{}, error) {
	fmt.Println("token from access token cookie: ", auth)

	id, err := a.JwtUsecase.ParseTokenAndGetID(auth)
	fmt.Println("ID from middlware", id, err)
	if err != nil {
		return nil, err
	}
	if !a.JwtUsecase.FindToken(id, auth) {
		fmt.Println("print if token not found from redis")
		return nil, err
	}
	role, err := a.JwtUsecase.ParseTokenAndGetRole(auth)
	fmt.Println("role from token: ", role, err)
	if err != nil {
		return nil, err
	}
	info := domain.User{
		ID:   id,
		Role: role,
	}
	return info, nil
}

func (a *Authorization) SetHeaders(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Response().Header().Set("Access-Control-Allow-Origin", "*")
		c.Response().Header().Set("Content-Type", "text/html")
		next(c)
		return nil
	}
}
