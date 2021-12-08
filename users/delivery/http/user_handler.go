package http

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
	"transaction-service/domain"
	config "transaction-service/users/delivery/http/middleware"

	utils "transaction-service/utils"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type UserHandler struct {
	UserUsecase domain.UserUsecase
	JwtUsecase  domain.JwtTokenUsecase
}

func NewUserHandler(e *echo.Echo, us domain.UserUsecase, jwt domain.JwtTokenUsecase) {

	handler := &UserHandler{UserUsecase: us, JwtUsecase: jwt}
	midd := config.InitAuthorization(jwt)

	e.GET("/", handler.Home, middleware.JWTWithConfig(midd.GetConfig()))
	e.GET("/login", handler.LoginPage).Name = "userSignInForm"
	e.POST("/login", handler.Signin)

	e.GET("/signup", handler.RegistrationPage)
	e.POST("/signup", handler.Registration, midd.SetHeaders)
	e.POST("/signup/admin", handler.AdminRegistration, middleware.JWTWithConfig(midd.GetConfig()))

	infoGroup := e.Group("/info")
	infoGroup.Use(middleware.JWTWithConfig(midd.GetConfig()))

	infoGroup.GET("", handler.GetAllUserInfo)
	infoGroup.GET("/:id", handler.GetUserInfo)

}

func (u *UserHandler) Home(e echo.Context) error {
	return e.NoContent(http.StatusOK)
}

func (u *UserHandler) Signin(e echo.Context) error {

	creds, err := u.ExtractBody(e)
	if err != nil {
		return e.String(http.StatusBadRequest, err.Error())
	}

	user, err := u.UserUsecase.GetUserByNameUsecase(creds.Username)
	if err != nil {
		return e.String(http.StatusForbidden, fmt.Sprintf("username is incorrect: %v", err))
	}
	if !utils.ComparePasswordHash(user.Password, creds.Password) {
		return e.String(http.StatusForbidden, "password is incorrect: %v")
	}

	signedToken, err := u.JwtUsecase.GenerateToken(user.ID, user.Role)
	if err != nil {
		return e.String(http.StatusInternalServerError, fmt.Sprintf("coundn't create token. Error: %v", err))
	}

	if err := u.JwtUsecase.InsertToken(user.ID, signedToken); err != nil {
		return e.String(http.StatusInternalServerError, fmt.Sprintf("coundn't insert token to redis. Error: %v", err))
	}

	u.SetCookie(e, signedToken)

	return e.String(http.StatusOK, fmt.Sprintf("Token: %s", signedToken))
}

func (u *UserHandler) SetCookie(e echo.Context, signedToken string) {
	ttl := u.JwtUsecase.GetAccessTTL()
	cookie := &http.Cookie{
		Name:    "access-token",
		Value:   signedToken,
		Expires: time.Now().Add(ttl),
	}
	e.SetCookie(cookie)
}

func (u *UserHandler) Registration(e echo.Context) error {

	userInfo, err := u.ExtractBody(e)
	if err != nil {
		return e.String(http.StatusInternalServerError, err.Error())
	}
	if userInfo.Role == "admin" {
		e.Redirect(http.StatusMovedPermanently, e.Echo().Reverse("userSignInForm"))
	}
	if err != nil {
		return e.String(http.StatusInternalServerError, err.Error())
	}
	if err := u.UserUsecase.CreateUserUsecase(userInfo); err != nil {
		return e.String(http.StatusBadRequest, err.Error())
	}
	return nil
}

func (u *UserHandler) AdminRegistration(e echo.Context) error {
	userInfo, err := u.ExtractBody(e)

	if err != nil {
		return e.String(http.StatusInternalServerError, err.Error())
	}
	meta, ok := e.Get("user").(map[int64]string)
	if !ok {
		return e.String(http.StatusInternalServerError, "cannot get meta info")
	}

	for _, role := range meta {
		if role != "admin" {
			return e.String(http.StatusForbidden, "access denied")
		}
	}
	if err := u.UserUsecase.CreateUserUsecase(userInfo); err != nil {
		return e.String(http.StatusBadRequest, err.Error())
	}
	return nil
}

func (u *UserHandler) ExtractBody(ctx echo.Context) (*domain.User, error) {
	user := &domain.User{}
	if err := ctx.Bind(user); err != nil {
		return nil, fmt.Errorf("bind body err: %w", err)
	}

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

	meta, ok := e.Get("user").(map[int64]string)
	if !ok {
		return e.String(http.StatusInternalServerError, "cannot get meta info")
	}

	for id, role := range meta {
		if role != "admin" && id != int64(newID) {
			return e.String(http.StatusForbidden, "access denied")
		}
	}

	user, err := u.UserUsecase.GetUserByIDUsecase(int64(newID))
	if err != nil {
		return e.String(http.StatusBadRequest, fmt.Sprintf("user not found: %v", err))
	}

	account, err := GetAccountInfo(e, *user, "http://localhost:8181/info/"+user.IIN)
	if err != nil {
		return e.String(http.StatusInternalServerError, fmt.Sprintf("get acc info err: %v", err))
	}

	return e.JSON(http.StatusOK, account)
}

func (u *UserHandler) GetAllUserInfo(e echo.Context) error {
	meta, ok := e.Get("user").(map[int64]string)
	if !ok {
		return e.String(http.StatusInternalServerError, "cannot get meta info")
	}

	for _, role := range meta {
		if role != "admin" {
			return e.String(http.StatusForbidden, "access denied")
		}
	}

	users, err := u.UserUsecase.GetAllUsecase()
	if err != nil {
		return e.String(http.StatusInternalServerError, fmt.Sprintf("%v", err))
	}

	// account := domain.Accounts{}
	all := []domain.Accounts{}

	for _, user := range users {

		acc, err := GetAccountInfo(e, user, "http://localhost:8181/info/"+user.IIN)
		if err != nil {
			return e.String(http.StatusInternalServerError, fmt.Sprintf("get account info error: %v", err))
		}
		all = append(all, acc...)
	}
	return e.JSON(http.StatusOK, all)
}

func GetAccountInfo(e echo.Context, user domain.User, url string) ([]domain.Accounts, error) {
	account := domain.Accounts{}
	all := []domain.Accounts{}

	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	resp, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	res.Body.Close()
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("get account error: StatusCode not 200")
	}

	if err := json.Unmarshal(resp, &all); err != nil {
		return nil, err
	}
	if len(all) == 0 {
		account.User = user
		all = append(all, account)
		return all, nil
	}
	for _, a := range all {
		a.User = user
		all = append(all, a)
	}
	return all, nil
}
