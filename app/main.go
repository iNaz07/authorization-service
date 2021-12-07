package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"
	"transaction-service/domain"
	_handler "transaction-service/users/delivery/http"
	_repo "transaction-service/users/repository/postgres"
	_usecase "transaction-service/users/usecase"

	"github.com/go-redis/redis"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/labstack/echo/v4"
)

//TODO: move to env
const (
	username = "postgres"
	password = "password"
	hostname = "localhost" //change as in docker-compose
	port     = 5432
	dbname   = "transaction" //change as in init.sql
)

func main() {

	client := connectRedis()
	db := connectDB()
	defer db.Close()

	token := domain.JwtToken{
		AccessSecret: "super secret code",
		RedisConn:    client,
		AccessTtl:    30 * time.Minute,
	}

	e := echo.New()

	userRepo := _repo.NewUserRepository(db)
	userUsecase := _usecase.NewUserUseCase(userRepo)
	jwtUsecase := _usecase.NewJWTUseCase(token)

	_handler.NewUserHandler(e, userUsecase, jwtUsecase)

	// midd := _middl.InitAuthorization(jwtUsecase, token)
	// infoGroup := e.Group("/info")
	// infoGroup.Use(middleware.JWTWithConfig(midd.GetConfig))
	// _handler.NewUserHandler(e, userUsecase, jwtUsecase)

	err := e.Start("127.0.0.1:8080")
	if err != nil && err != http.ErrServerClosed {
		log.Fatal(`shutting down the server`, err) //better do not fatal, cuz defer wouldn't done. think bout it
	}
}

func connectRedis() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	_, err := client.Ping().Result()
	if err != nil {
		log.Fatalf("Ping error: %v", err)
	}
	return client
}

func connectDB() *pgxpool.Pool {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s", username, password, hostname, port, dbname)

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		log.Fatalf("Open db error: %v", err)
	}
	// should read bout them
	// config.MaxConns = 25
	// config.MaxConnLifetime = 5 * time.Minute
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
	defer cancel()
	db, err := pgxpool.ConnectConfig(ctx, config)
	if err != nil {
		log.Fatalf("Connect db error: %v", err)
	}

	if err := db.Ping(ctx); err != nil {
		log.Fatalf("Ping db error: %v", err)
	}
	_, err = db.Exec(ctx, `
	DROP TABLE users;
	`)
	if err != nil {
		log.Fatalf("Drop table error: %v", err)
	}
	_, err = db.Exec(ctx, `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		username VARCHAR (100) NOT NULL UNIQUE,
		password TEXT NOT NULL,
		iin VARCHAR (255) NOT NULL,
		role VARCHAR (24) NOT NULL,
		registerDate VARCHAR (255) NOT NULL
	);
	`)
	if err != nil {
		log.Fatalf("Create table error: %v", err)
	}
	_, err = db.Exec(ctx,
		`INSERT INTO users(username, password, iin, role, registerDate) VALUES ($1, $2, $3, $4, $5)`,
		"admin", "pass", "940217200216", "admin", time.Now().Format("2006-01-02 15:04:05"))
	if err != nil {
		log.Fatalf("Add admin error: %v", err)
	}
	return db
}
