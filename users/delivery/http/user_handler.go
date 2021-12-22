package http

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"

	"net/http"
	"strconv"
	"time"
	"transaction-service/domain"
	config "transaction-service/users/delivery/http/middleware"

	utils "transaction-service/utils"

	"github.com/rs/zerolog/log"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type UserHandler struct {
	UserUsecase domain.UserUsecase
	JwtUsecase  domain.JwtTokenUsecase
}

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func NewUserHandler(e *echo.Echo, us domain.UserUsecase, jwt domain.JwtTokenUsecase) {
	t := &Template{
		templates: template.Must(template.ParseGlob("../templates/*.html")),
	}
	e.Renderer = t

	handler := &UserHandler{UserUsecase: us, JwtUsecase: jwt}
	midd := config.InitAuthorization(jwt)

	e.Use(midd.SetHeaders)

	e.GET("/login", handler.LoginPage).Name = "userSignInForm"
	e.POST("/login", handler.Signin)

	e.GET("/signup", handler.RegistrationPage)
	e.POST("/signup", handler.Registration)
	e.GET("/", handler.Home, middleware.JWTWithConfig(midd.GetConfig()))

	infoGroup := e.Group("/user")
	infoGroup.Use(middleware.JWTWithConfig(midd.GetConfig()))

	infoGroup.GET("/info/all", handler.GetAllUserInfo)
	infoGroup.GET("/info/:id", handler.GetUserInfo)
	infoGroup.GET("/upgrade/:username", handler.UpgradeRole)
	infoGroup.GET("/home", handler.Home)

}

func (u *UserHandler) Home(e echo.Context) error {
	meta, ok := e.Get("user").(domain.User)
	if !ok {
		log.Err(domain.ErrorMetaNotFound).Msg("unauthorized")
		return e.Render(http.StatusUnauthorized, "error.html", "access denied")
	}

	ctx := e.Request().Context()

	user, err := u.UserUsecase.GetUserByIDUsecase(ctx, meta.ID)
	if err != nil {
		logerr := err.(*domain.LogError)
		log.Err(logerr.Err).Msg(logerr.Message)
		return e.String(logerr.Code, "Access denied")
	}
	// return e.JSON(http.StatusOK, user)
	return e.Render(http.StatusOK, "home.html", user)
}

func (u *UserHandler) Signin(e echo.Context) error {

	creds := u.ExtractCreds(e)
	if creds.Username == "" || creds.Password == "" {
		log.Log().Msg("username or password must be filled")
		return e.Render(http.StatusBadRequest, "error.html", "username or password must be filled")
	}
	ctx := e.Request().Context()
	user, err := u.UserUsecase.GetUserByNameUsecase(ctx, creds.Username)
	if err != nil {
		logerr := err.(*domain.LogError)
		log.Err(logerr.Err).Msg(logerr.Message)
		return e.Render(logerr.Code, "error.html", "incorrect username")
	}
	if !utils.ComparePasswordHash(user.Password, creds.Password) {
		log.Log().Msg("incorrect password")
		return e.Render(http.StatusForbidden, "error.html", "incorrect password")
	}

	signedToken, err := u.JwtUsecase.GenerateToken(user.ID, user.Role, user.IIN)
	if err != nil {
		logerr := err.(*domain.LogError)
		log.Err(logerr.Err).Msg(logerr.Message)
		return e.Render(logerr.Code, "error.html", "Unexpected error. Please try again in several minutes")
		// return e.String(http.StatusInternalServerError, "generate token error")
	}

	if err := u.JwtUsecase.InsertToken(user.ID, signedToken); err != nil {
		logerr := err.(*domain.LogError)
		log.Err(logerr.Err).Msg(logerr.Message)
		return e.Render(logerr.Code, "error.html", "Unexpected error. Please try again in several minutes")
		// return e.String(http.StatusInternalServerError, "insert error")
	}

	u.SetCookie(e, signedToken)
	// return e.JSON(http.StatusOK, user)
	return e.Render(http.StatusOK, "home.html", user)
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

	userInfo := u.ExtractCreds(e)
	ctx := e.Request().Context()
	if err := u.UserUsecase.CreateUserUsecase(ctx, userInfo); err != nil {
		logerr := err.(*domain.LogError)
		log.Err(logerr.Err).Msg(logerr.Message)
		return e.Render(logerr.Code, "error.html", logerr.Message)
	}
	// return e.JSON(http.StatusCreated, "Successfully registered. Now you can log in")
	return e.Render(http.StatusCreated, "login.html", "Successfully registered. Now you can log in")
}

func (u *UserHandler) UpgradeRole(e echo.Context) error {

	username := e.Param("username")
	meta, ok := e.Get("user").(domain.User)
	if !ok {
		log.Err(domain.ErrorMetaNotFound).Msg("unauthorized")
		return e.Render(http.StatusUnauthorized, "error.html", "access denied")
	}

	if meta.Role != "admin" {
		log.Log().Msg("role not admin")
		return e.Render(http.StatusForbidden, "error.html", "access denied")
	}
	ctx := e.Request().Context()
	if err := u.UserUsecase.UpgradeUserUsecase(ctx, username); err != nil {
		logerr := err.(*domain.LogError)
		log.Err(logerr.Err).Msg(logerr.Message)
		return e.Render(http.StatusInternalServerError, "error.html", "Unexpected error. Please try again")
	}
	// return e.String(http.StatusOK, fmt.Sprintf("User %s upgraded to administrator", username))
	return e.Render(http.StatusOK, "error.html", fmt.Sprintf("User %s upgraded to administrator", username))
}

func (u *UserHandler) ExtractCreds(c echo.Context) *domain.User {
	return &domain.User{
		Username: c.FormValue("username"),
		Password: c.FormValue("password"),
		IIN:      c.FormValue("iin"),
		Role:     c.FormValue("role"),
	}
}

func (u *UserHandler) LoginPage(e echo.Context) error {
	return e.Render(http.StatusOK, "login.html", nil)
}

func (u *UserHandler) RegistrationPage(e echo.Context) error {
	return e.Render(http.StatusOK, "signup.html", nil)
}

func (u *UserHandler) GetUserInfo(e echo.Context) error {

	newID, err := strconv.Atoi(e.Param("id"))
	if err != nil {
		log.Err(err).Msg(err.Error())
		return e.Render(http.StatusBadRequest, "error.html", "Invalid ID")
	}

	meta, ok := e.Get("user").(domain.User)
	if !ok {
		log.Err(domain.ErrorMetaNotFound).Msg("unauthorized")
		return e.Render(http.StatusUnauthorized, "error.html", "access denied")
	}

	if meta.Role != "admin" && meta.ID != int64(newID) {
		log.Log().Msg("requesting confidentional information")
		return e.Render(http.StatusForbidden, "error.html", "access denied")
	}
	ctx := e.Request().Context()
	user, err1 := u.UserUsecase.GetUserByIDUsecase(ctx, int64(newID))
	if err1 != nil {
		logerr := err1.(*domain.LogError)
		log.Err(logerr.Err).Msg(logerr.Message)
		// return e.String(http.StatusBadRequest, fmt.Sprintf("user not found: %v", err)) logg
		return e.Render(http.StatusBadRequest, "error.html", logerr.Message)
	}
	acc, err2 := GetAccountInfo(e, user.IIN)
	if err2 != nil {
		logerr := err2.(*domain.LogError)
		log.Err(logerr).Msg(logerr.Message)
		info := domain.UserInfo{
			User: *user,
		}
		return e.Render(http.StatusOK, "userinfo.html", info)
	}
	info := domain.UserInfo{
		User:     *user,
		Accounts: acc,
	}
	// return e.JSON(http.StatusOK, info)
	return e.Render(http.StatusOK, "userinfo.html", info)
}

func (u *UserHandler) GetAllUserInfo(e echo.Context) error {

	meta, ok := e.Get("user").(domain.User)
	if !ok {
		log.Err(domain.ErrorMetaNotFound).Msg("unauthorized")
		return e.Render(http.StatusUnauthorized, "error.html", "access denied")
	}

	if meta.Role != "admin" {
		log.Log().Msg("requesting confidentional information")
		return e.Render(http.StatusForbidden, "error.html", "access denied")
	}
	ctx := e.Request().Context()
	users, err := u.UserUsecase.GetAllUsecase(ctx)
	if err != nil {
		logerr := err.(*domain.LogError)
		log.Err(logerr.Err).Msg(logerr.Message)
		return e.Render(logerr.Code, "error.html", "Unexpected error. Please try again")
	}

	all := []domain.UserInfo{}

	for _, user := range users {
		acc, err1 := GetAccountInfo(e, user.IIN)
		if err1 != nil {
			logErr := err1.(*domain.LogError)
			log.Err(logErr).Msg(logErr.Message)
			info := domain.UserInfo{
				User: user,
			}
			all = append(all, info)
			continue
			// return e.Render(http.StatusOK, "alluser.html", all)
		}

		info := domain.UserInfo{
			User:     user,
			Accounts: acc,
		}
		all = append(all, info)
	}
	return e.Render(http.StatusOK, "alluser.html", all)
	// return e.JSON(http.StatusOK, all)
}

func GetAccountInfo(e echo.Context, iin string) ([]domain.Accounts, error) {
	all := []domain.Accounts{}

	cookie, err := e.Cookie("access-token")
	if err != nil {
		return nil, &domain.LogError{"cookie not found", err, http.StatusUnauthorized}
	}

	req, err := http.NewRequest("GET", "http://localhost:8181/account/info/"+iin+"/auth", nil)
	if err != nil {
		return nil, &domain.LogError{"create new request error", err, http.StatusInternalServerError}
	}

	req.AddCookie(cookie)
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, &domain.LogError{"send request error", err, http.StatusInternalServerError}
	}

	resp, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, &domain.LogError{"read response body error", err, http.StatusInternalServerError}
	}
	res.Body.Close()

	if res.StatusCode != 200 {
		return nil, &domain.LogError{"accounts not found", err, res.StatusCode}
	}
	if err := json.Unmarshal(resp, &all); err != nil {
		log.Log().Msg(string(resp))
		return nil, &domain.LogError{"unmarshal response body error", err, http.StatusInternalServerError}
	}
	return all, nil
}
