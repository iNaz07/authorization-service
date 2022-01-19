package main

import (
	"context"
	"fmt"

	"net/http"
	"time"
	"transaction-service/domain"
	_handler "transaction-service/users/delivery/http"
	_repo "transaction-service/users/repository/postgres"
	_redis "transaction-service/users/repository/redis"
	_usecase "transaction-service/users/usecase"

	"github.com/rs/zerolog/log"

	"github.com/go-redis/redis"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"
)

func init() {
	viper.SetConfigFile(`config.json`)
	if err := viper.ReadInConfig(); err != nil {
		log.Fatal().Err(err).Msg("get configuration error")
	}
}

func main() {

	client := connectRedis()

	token := domain.JwtToken{
		AccessSecret: viper.GetString(`token.secret`),
		AccessTtl:    viper.GetDuration(`token.ttl`) * time.Minute,
	}
	redis := _redis.NewRedisRepo(client)
	timeout := viper.GetDuration(`timeout`) * time.Second

	db := connectDB()
	defer db.Close()

	userRepo := _repo.NewUserRepository(db)
	userUsecase := _usecase.NewUserUseCase(userRepo, timeout)
	jwtUsecase := _usecase.NewJWTUseCase(token, redis)

	e := echo.New()
	_handler.NewUserHandler(e, userUsecase, jwtUsecase)

	err := e.Start(viper.GetString(`addr`))
	if err != nil && err != http.ErrServerClosed {
		log.Fatal().Err(err).Msg(`shutting down the server`)
	}
}

func connectRedis() *redis.Client {

	address := viper.GetString(`redis.address`)
	password := viper.GetString(`redis.password`)

	client := redis.NewClient(&redis.Options{
		Addr:     address,
		Password: password,
		DB:       0,
	})

	_, err := client.Ping().Result()
	if err != nil {
		log.Fatal().Err(err).Msg("redis ping error")
	}
	return client
}

func connectDB() *pgxpool.Pool {

	username := viper.GetString(`postgres.user`)
	password := viper.GetString(`postgres.password`)
	hostname := viper.GetString(`postgres.host`)
	port := viper.GetInt(`postgres.port`)
	dbname := viper.GetString(`postgres.dbname`)

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s", username, password, hostname, port, dbname)

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		log.Fatal().Err(err).Msg("parse dsn config error")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
	defer cancel()
	db, err := pgxpool.ConnectConfig(ctx, config)
	if err != nil {
		log.Fatal().Err(err).Msg("connect postgres error")
	}

	if err := db.Ping(ctx); err != nil {
		log.Fatal().Err(err).Msg("db ping error")
	}
	//TODO: move all to init.sql
	//temporary

	// _, err = db.Exec(ctx, `
	// DROP TABLE users;
	// `)
	// if err != nil {
	// 	log.Fatalf("Drop table error: %v", err)
	// }
	_, err = db.Exec(ctx, `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY UNIQUE,
		username VARCHAR (255) NOT NULL UNIQUE,
		password TEXT NOT NULL,
		iin VARCHAR (255) NOT NULL UNIQUE,
		role VARCHAR (24) NOT NULL,
		registerDate TEXT NOT NULL
	);
	`)
	if err != nil {
		log.Fatal().Err(err).Msg("Create table error")
	}
	_, err = db.Exec(ctx,
		`INSERT INTO users(username, password, iin, role, registerDate) VALUES ($1, $2, $3, $4, $5)`,
		"admin", "pass", "940217200216", "admin", time.Now().Format("2006-01-02 15:04:05"))
	if err != nil {
		log.Printf("Admin already exist: %v", err)
	}
	return db
}
