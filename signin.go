package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"
)

func (a *Admin) Signin(eCtx echo.Context) error {

	login, password := a.ExtractCredential(eCtx)

	if login != a.Username || password != a.Password {
		return eCtx.String(http.StatusForbidden, "invalid credentials")
	}

	accessTokenClaims := jwt.MapClaims{}
	accessTokenClaims["id"] = a.Id
	accessTokenClaims["iat"] = time.Now().Unix()
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessTokenClaims)

	signedToken, err := accessToken.SignedString([]byte("super secret word"))
	if err != nil {
		return eCtx.String(http.StatusInternalServerError, fmt.Sprintf("coundn't create token. Error: %v", err))
	}

	if err := a.InsertToken(signedToken); err != nil {
		return eCtx.String(http.StatusInternalServerError, fmt.Sprintf("coundn't insert token in redis. Error: %v", err))
	}

	return eCtx.String(http.StatusOK, fmt.Sprintf("Token: %s", signedToken))
}

func (a *Admin) ExtractCredential(ctx echo.Context) (login string, pass string) {
	return ctx.FormValue("login"), ctx.FormValue("password")
}

func (a *Admin) ExtractToken(ctx echo.Context) (token string, err error) {
	header := ctx.Request().Header.Get("Authorization")
	if header == "" {
		return "", fmt.Errorf("authorization header not found")
	}
	parsedHeader := strings.Split(header, " ")
	if len(parsedHeader) != 2 || parsedHeader[0] != "Bearer" {
		return "", fmt.Errorf("invalid authorization header")
	}
	token = parsedHeader[1]
	return token, nil
}

func (a *Admin) ParseToken(token string) (int64, error) {
	JWTToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("failed to extract token metadata, unexpected signing method: %v", token.Header["alg"])
		}
		return []byte("super secret word"), nil
	})

	if err != nil {
		return 0, err
	}

	claims, ok := JWTToken.Claims.(jwt.MapClaims)
	var userId float64

	if ok && JWTToken.Valid {
		userId, ok = claims["id"].(float64)
		if !ok {
			return 0, fmt.Errorf("field id not found")
		}
		return int64(userId), nil
	}
	return 0, fmt.Errorf("Invalid token")
}

func (a *Admin) FindToken(token string) bool {
	key := fmt.Sprintf("user:%d", a.Id)

	value, err := a.redisConn.Get(key).Result()
	if err != nil {
		return false
	}
	return token == value
}

func (a *Admin) InsertToken(token string) error {
	key := fmt.Sprintf("user:%d", a.Id)
	return a.redisConn.Set(key, token, 20*time.Second).Err()
}
