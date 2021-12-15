package http

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
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
	infoGroup.POST("/upgrade/:username", handler.UpgradeRole)
	infoGroup.GET("/home", handler.Home)

}

func (u *UserHandler) Home(e echo.Context) error {
	meta, ok := e.Get("user").(domain.User)
	if !ok {
		return e.String(http.StatusInternalServerError, "cannot get meta info")
	}

	user, err := u.UserUsecase.GetUserByIDUsecase(meta.ID)
	if err != nil {
		fmt.Println(err.Error()) //log
		return e.String(http.StatusForbidden, "Access denied")
	}

	return e.Render(http.StatusOK, "home.html", user)
}

func (u *UserHandler) Signin(e echo.Context) error {

	creds := u.ExtractCreds(e)
	if creds.Username == "" || creds.Password == "" {
		return e.Render(http.StatusBadRequest, "error.html", "username or password must be filled")
	}
	user, err := u.UserUsecase.GetUserByNameUsecase(creds.Username)
	if err != nil {
		e.String(http.StatusForbidden, fmt.Sprintf("username is incorrect: %v", err)) //to logs
		return e.Render(http.StatusForbidden, "error.html", "username is incorrect")
	}
	if !utils.ComparePasswordHash(user.Password, creds.Password) {
		fmt.Println("compare passwords", user.Password, creds.Password)
		e.String(http.StatusForbidden, "password is incorrect") //to logs
		return e.Render(http.StatusForbidden, "error.html", "password is incorrect")
	}

	signedToken, err := u.JwtUsecase.GenerateToken(user.ID, user.Role, user.IIN)
	if err != nil {
		fmt.Println("error when generating new token: ", err) //to log
		return e.Render(http.StatusInternalServerError, "error.html", "Unexpected error. Please try again in several minutes")
	}

	if err := u.JwtUsecase.InsertToken(user.ID, signedToken); err != nil {
		fmt.Println("error when inserting token into redis", err) //log
		return e.Render(http.StatusInternalServerError, "error.html", "Unexpected error. Please try again in several minutes")
	}

	u.SetCookie(e, signedToken)
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
	fmt.Println("user data from client", userInfo)

	if err := u.UserUsecase.CreateUserUsecase(userInfo); err != nil {
		// return e.String(http.StatusBadRequest, err.Error())      // to logs
		return e.Render(http.StatusBadRequest, "error.html", err.Error())
	}
	return e.Render(http.StatusOK, "login.html", "Successfully registered. Now you can log in")
}

//no need ?
func (u *UserHandler) UpgradeRole(e echo.Context) error {

	username := e.Param("username")
	meta, ok := e.Get("user").(domain.User)
	if !ok {
		return e.String(http.StatusForbidden, "Access denied. Please authorize")
	}

	if meta.Role != "admin" {
		return e.String(http.StatusForbidden, "Access denied")
	}

	if err := u.UserUsecase.UpgradeUserUsecase(username); err != nil {
		log.Printf("upgrade error: %v", err)
		return e.Render(http.StatusInternalServerError, "error.html", "Unexpected error. Please try again")
	}
	return e.Render(http.StatusOK, "error.html", fmt.Sprintf("User %s upgraded to administrator"))
}

//no need
func (u *UserHandler) checkAuth(e echo.Context) bool {

	meta, ok := e.Get("user").(map[int64]string)
	if !ok {
		return false
	}

	for _, role := range meta {
		if role != "admin" {
			return false
		}
	}
	return true
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
		fmt.Println(err.Error()) //log
		return e.Render(http.StatusBadRequest, "error.html", "Invalid ID")
	}

	meta, ok := e.Get("user").(domain.User)
	if !ok {
		return e.String(http.StatusForbidden, "Access denied. Please, authorize")
	}

	if meta.Role != "admin" && meta.ID != int64(newID) {
		return e.String(http.StatusForbidden, "Access denied.")
	}

	user, err := u.UserUsecase.GetUserByIDUsecase(int64(newID))
	if err != nil {
		// return e.String(http.StatusBadRequest, fmt.Sprintf("user not found: %v", err)) logg
		return e.Render(http.StatusBadRequest, "error.html", "user not found")
	}

	acc, err := GetAccountInfo(e, user.IIN)
	if err != nil {
		return err
	}
	info := domain.UserInfo{
		User:     *user,
		Accounts: acc,
	}

	return e.Render(http.StatusOK, "userinfo.html", info)
}

func (u *UserHandler) GetAllUserInfo(e echo.Context) error {

	meta, ok := e.Get("user").(domain.User)
	if !ok {
		fmt.Println("cannot get meta info") //log
		return e.String(http.StatusForbidden, "Access denied. Please, authorize")
	}

	if meta.Role != "admin" {
		return e.String(http.StatusForbidden, "Access denied.")
	}

	users, err := u.UserUsecase.GetAllUsecase()
	if err != nil {
		e.String(http.StatusInternalServerError, fmt.Sprintf("%v", err)) //logg
		return e.Render(http.StatusInternalServerError, "error.html", "Unexpected error. Please try again")
	}

	all := []domain.UserInfo{}

	for _, user := range users {
		acc, err := GetAccountInfo(e, user.IIN)
		if err != nil {
			return err
		}

		info := domain.UserInfo{
			User:     user,
			Accounts: acc,
		}
		all = append(all, info)
	}
	return e.Render(http.StatusOK, "alluser.html", all)
}

func GetAccountInfo(e echo.Context, iin string) ([]domain.Accounts, error) {
	all := []domain.Accounts{}

	cookie, err := e.Cookie("access-token")
	if err != nil {
		return nil, e.String(http.StatusForbidden, err.Error())
	}

	req, err := http.NewRequest("GET", "http://localhost:8181/account/info/"+iin, nil)
	if err != nil {
		fmt.Println(err.Error())
		return nil, e.String(http.StatusInternalServerError, "Unexpected error. Please try again")
	}

	req.AddCookie(cookie)
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, e.String(http.StatusInternalServerError, err.Error())
	}

	resp, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, e.String(http.StatusInternalServerError, err.Error())
	}
	res.Body.Close()

	if res.StatusCode != 200 {
		fmt.Println("no accounts", res.StatusCode, string(resp)) //log
		return nil, nil
	}
	if err := json.Unmarshal(resp, &all); err != nil {
		return nil, e.String(http.StatusInternalServerError, err.Error())
	}
	return all, nil
}
