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
		TokenLookup: "cookie:access-token",
		ParseTokenFunc: func(auth string, c echo.Context) (interface{}, error) {
			fmt.Println("token from access token cookie: ", auth)
			info := make(map[int64]string)
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
			info[id] = role
			return info, nil

		},
		ErrorHandlerWithContext: a.JwtUsecase.JWTErrorChecker,
	}
}

// func (a *Authorization) CheckAuth(next echo.HandlerFunc) echo.HandlerFunc {
// 	return func(c echo.Context) error {
// 		cookie, err := c.Cookie("access_token")
// 		if err != nil {
// 			return c.String(http.StatusForbidden, fmt.Sprintf("cookie not found: %v", err))
// 		}
// 		id, err := a.JwtUsecase.ParseTokenAndGetID(cookie.Value)
// 		if err != nil {
// 			return c.String(http.StatusInternalServerError, fmt.Sprintf("parse token error: %v", err))
// 		}
// 		if !a.JwtUsecase.FindToken(id, cookie.Value) {
// 			return c.String(http.StatusForbidden, fmt.Sprintf("invalid token: %v", err))
// 		}
// 		role, err := a.JwtUsecase.ParseTokenAndGetRole(cookie.Value)
// 		if err != nil {
// 			return c.String(http.StatusInternalServerError, fmt.Sprintf("parse token error: %v", err))
// 		}
// 		key := strconv.Itoa(int(id))
// 		// type value string
// 		// c.Set(key, value(role))
// 		c.Set(key, role)
// 		next(c)
// 		return nil
// 	}
// }
