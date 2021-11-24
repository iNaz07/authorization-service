package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"transaction-service/domain"

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
}

func (u *UserHandler) Signin(e echo.Context) error {
	//TODO: get from body
	login, password := u.ExtractCreds(e)
	fmt.Println("from signin", login, password)
	user, err := u.UserUsecase.GetUserByNameUsecase(login)
	if err != nil || user.Password != password {
		fmt.Println("from signin checking user from db", err)
		return e.String(http.StatusForbidden, "invalid credentials")
	}
	signedToken, err := u.JwtUsecase.GenerateToken(user.ID, user.Role)
	if err != nil {
		return e.String(http.StatusInternalServerError, fmt.Sprintf("coundn't create token. Error: %v", err))
	}

	if err := u.JwtUsecase.InsertToken(user.ID, signedToken); err != nil {
		return e.String(http.StatusInternalServerError, fmt.Sprintf("coundn't insert token in redis. Error: %v", err))
	}
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
