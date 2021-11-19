package main

import (
	"log"
	"net/http"
	"github.com/go-redis/redis"
	"github.com/labstack/echo/v4"
)

type Admin struct {
	Id       int64
	Username string
	Password string
	redisConn *redis.Client
}

func main() {
	e := echo.New()

	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	pong, err := client.Ping().Result()
	if err != nil {
		log.Fatalf("Ping error: %v", err)
	}
	log.Println(pong)

	admin := &Admin{
		Id:       1,
		Username: "admin",
		Password: "password",
		redisConn: client,
	}
	e.GET("/login", admin.loginPage)
	e.POST("/login", admin.Signin)
	e.GET("/signup", admin.registrationPage)
	e.POST("/signup", admin.Registration)
	e.GET("/", func(c echo.Context) error {
		token, err := admin.ExtractToken(c)
		if err != nil {
			log.Printf("Extract token error: %v", err)
			return c.Redirect(http.StatusMovedPermanently, "/login")
		}

		id, err := admin.ParseToken(token)
		if err != nil {
			log.Printf("Parse token error: %v", err)
			return c.Redirect(http.StatusMovedPermanently, "/login")
		}

		ok := admin.FindToken(token)
		if !ok {
			log.Printf("Getting token failed")
			return c.Redirect(http.StatusMovedPermanently, "/login")
		}

		c.Set("userID", id)
		return c.String(http.StatusOK, "authorization success")
	})

	err = e.Start("127.0.0.1:8080")
	if err != nil && err != http.ErrServerClosed {
		log.Fatal(`shutting down the server`, err)
	}
}

func (a *Admin) loginPage(eCtx echo.Context) error {
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

func (a *Admin) registrationPage(eCtx echo.Context) error {
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
	<label for="firstname">Firstname:</label> <br>
	<input type="text" id="firstname" name="firstname"> <br>
	<label for="lastname">Lastname:</label> <br>
	<input type="text" id="lastname" name="lastname"> <br>
	<label for="age">Age:</label> <br>
	<input type="text" id="age" name="age"> <br>
	<input type="submit" value="Sign in">
</form>
</body>
</html>
	`)
}
