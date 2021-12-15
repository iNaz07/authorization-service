package middleware

import (
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

	id, err := a.JwtUsecase.ParseTokenAndGetID(auth)
	if err != nil {
		return nil, err
	}
	if !a.JwtUsecase.FindToken(id, auth) {
		return nil, err
	}
	role, err := a.JwtUsecase.ParseTokenAndGetRole(auth)
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
