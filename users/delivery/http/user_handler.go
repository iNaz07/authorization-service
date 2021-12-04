package http

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
	"transaction-service/domain"
	utils "transaction-service/utils"

	"github.com/labstack/echo/v4"
)

type UserHandler struct {
	UserUsecase domain.UserUsecase
	JwtUsecase  domain.JwtTokenUsecase
}

func NewUserHandler(e *echo.Echo, us domain.UserUsecase, jwt domain.JwtTokenUsecase) {
	handler := &UserHandler{UserUsecase: us, JwtUsecase: jwt}
	e.GET("/login", handler.LoginPage)
	e.POST("/login", handler.Signin)
	e.POST("/signup", handler.Registration)
	e.GET("/signup", handler.RegistrationPage)
	e.GET("/info/:id", handler.GetUserInfo)
	e.GET("/info", handler.GetAllUserInfo)
	// e.Use(middleware.JWTWithConfig(config))
}

func (u *UserHandler) Signin(e echo.Context) error {
	//TODO: get from body
	login, password := u.ExtractCreds(e)
	fmt.Println("from signin", login, password)
	user, err := u.UserUsecase.GetUserByNameUsecase(login)
	if err != nil {
		fmt.Println("from signin checking user from db", err)
		return e.String(http.StatusForbidden, fmt.Sprintf("username is incorrect: %v", err))
	}
	if !utils.ComparePasswordHash(user.Password, password) {
		return e.String(http.StatusForbidden, "password is incorrect: %v")
	}

	signedToken, err := u.JwtUsecase.GenerateToken(user.ID, user.Role)
	if err != nil {
		return e.String(http.StatusInternalServerError, fmt.Sprintf("coundn't create token. Error: %v", err))
	}

	if err := u.JwtUsecase.InsertToken(user.ID, signedToken); err != nil {
		return e.String(http.StatusInternalServerError, fmt.Sprintf("coundn't insert token in redis. Error: %v", err))
	}
	ttl := u.JwtUsecase.GetAccessTTL()
	cookie := new(http.Cookie)
	cookie.Name = "access_token"
	cookie.Value = signedToken
	cookie.Expires = time.Now().Add(ttl)
	e.SetCookie(cookie)
	return e.String(http.StatusOK, fmt.Sprintf("Token: %s", signedToken))
}

func (u *UserHandler) ExtractCreds(ctx echo.Context) (login string, pass string) {
	return ctx.FormValue("login"), ctx.FormValue("password")
}

func (u *UserHandler) Registration(e echo.Context) error {
	userInfo, err := u.ExtractBody(e)
	fmt.Println("from registration", userInfo, err)
	if err != nil {
		return e.String(http.StatusInternalServerError, err.Error())
	}
	if err := u.UserUsecase.CreateUserUsecase(userInfo); err != nil {
		return e.String(http.StatusBadRequest, err.Error())
	}
	return nil
}

func (u *UserHandler) ExtractBody(ctx echo.Context) (*domain.User, error) {
	var user *domain.User
	fmt.Println("body is", ctx.Request().Body)
	if err := json.NewDecoder(ctx.Request().Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("unmarshal body err: %w", err)
	}
	//if checkbox admin true {register as admin}
	if user.Role == "" {
		user.Role = "user"
	}
	return user, nil
}

func (u *UserHandler) LoginPage(eCtx echo.Context) error {
	return eCtx.HTML(http.StatusOK, `
	<html>
<head>
</head>
<body>
<form action="/login" method="post">
	<label for="login">Login:</label> <br>
	<input type="text" id="login" name="login"> <br>
	<label for="password">Password:</label> <br>
	<input type="text" id="password" name="password"> <br>
	<input type="submit" value="Sign in">
</form>
</body>
</html>
	`)
}

func (u *UserHandler) RegistrationPage(e echo.Context) error {
	return e.HTML(http.StatusOK, `
	<html>
<head>
</head>
<body>
<form action="/login" method="post">
	<label for="login">Login:</label> <br>
	<input type="text" id="login" name="login"> <br>
	<label for="password">Password:</label> <br>
	<input type="text" id="password" name="password"> <br>
	<label for="iin">IIN:</label> <br>
	<input type="text" id="iin" name="iin"> <br>
	<label for="role">Role:</label> <br>
	<input type="text" id="role" name="role"> <br>
	<input type="submit" value="Sign up">
</form>
</body>
</html>
	`)
}

func (u *UserHandler) GetUserInfo(e echo.Context) error {
	newID, err := strconv.Atoi(e.Param("id"))
	if err != nil {
		return e.String(http.StatusBadRequest, "Invalid ID")
	}

	cookie, err := e.Cookie("access_token")
	if err != nil {
		return e.String(http.StatusForbidden, fmt.Sprintf("cookie not found: %v", err))
	}
	role, err := u.JwtUsecase.ParseTokenAndGetRole(cookie.Value)
	if err != nil {
		return e.String(http.StatusInternalServerError, fmt.Sprintf("parse token error: %v", err))
	}
	if role != "admin" {
		id, err := u.JwtUsecase.ParseTokenAndGetID(cookie.Value)
		// log.Println("id and cookie from auth service", id, cookie)
		if err != nil {
			return e.String(http.StatusInternalServerError, fmt.Sprintf("parse token error: %v", err))
		}
		if id != int64(newID) {
			return e.String(http.StatusForbidden, "access denied")
		}
	}

	user, err := u.UserUsecase.GetUserByIDUsecase(int64(newID))
	if err != nil {
		return e.String(http.StatusBadRequest, fmt.Sprintf("user not found: %v", err))
	}

	account := struct {
		User         domain.User
		Number       string `json:"number"`
		Balance      int64  `json:"balance"`
		RegisterDate string `json:"registerDate"`
	}{}
	res, err := http.Get("http://localhost:8181/account/info/" + user.IIN)
	// log.Println("resp from get req from ts:  ", res, "ERROR is: ", err)
	if err != nil {
		return e.String(http.StatusInternalServerError, fmt.Sprintf("get user accounts error: %v", err))
	}
	resp, err := io.ReadAll(res.Body)
	if err != nil {
		return e.String(http.StatusInternalServerError, fmt.Sprintf("read body error: %v", err))
	}
	res.Body.Close()
	if res.StatusCode != 200 {
		return e.String(res.StatusCode, fmt.Sprintf("get accounts info error: %v", string(resp)))
	}
	if err := json.Unmarshal(resp, &account); err != nil {
		return e.String(http.StatusInternalServerError, fmt.Sprintf("unmarshal responce err: %v", err))
	}
	account.User = *user

	return e.JSON(http.StatusOK, account)
}

func (u *UserHandler) GetAllUserInfo(e echo.Context) error {
	cookie, err := e.Cookie("access_token")
	if err != nil {
		return e.String(http.StatusForbidden, fmt.Sprintf("cookie not found: %v", err))
	}
	role, err := u.JwtUsecase.ParseTokenAndGetRole(cookie.Value)
	if err != nil {
		return e.String(http.StatusInternalServerError, fmt.Sprintf("parse token error: %v", err))
	}
	if role != "admin" {
		return e.String(http.StatusForbidden, fmt.Sprintf("access denied: %v", err))
	}
	user, err := u.UserUsecase.GetAllUsecase()
	if err != nil {
		return e.String(http.StatusInternalServerError, fmt.Sprintf("%v", err))
	}

	return e.JSON(http.StatusOK, user)
}
