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
	"github.com/spf13/viper"
)

func init() {
	viper.SetConfigFile("config.json")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatal("get configuration error: ", err)
	}
}

func main() {

	client := connectRedis()

	token := domain.JwtToken{
		AccessSecret: viper.GetString(`token.secret`),
		RedisConn:    client,
		AccessTtl:    viper.GetDuration(`token.ttl`) * time.Minute,
	}

	db := connectDB()
	defer db.Close()

	userRepo := _repo.NewUserRepository(db)
	userUsecase := _usecase.NewUserUseCase(userRepo)
	jwtUsecase := _usecase.NewJWTUseCase(token)

	e := echo.New()
	_handler.NewUserHandler(e, userUsecase, jwtUsecase)

	err := e.Start(viper.GetString(`addr`))
	if err != nil && err != http.ErrServerClosed {
		log.Fatal(`shutting down the server`, err) //better do not fatal, cuz defer wouldn't done. think bout it
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
		log.Fatalf("Ping error: %v", err)
	}
	return client
}

func connectDB() *pgxpool.Pool {

	username := viper.GetString(`postgres.user`)
	password := viper.GetString(`postgres.password`) //"password"
	hostname := viper.GetString(`postgres.host`)     //"localhost"
	port := viper.GetInt(`postgres.port`)            //5432
	dbname := viper.GetString(`postgres.dbname`)     //auth

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s", username, password, hostname, port, dbname)

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		log.Fatalf("Open db error: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
	defer cancel()
	db, err := pgxpool.ConnectConfig(ctx, config)
	if err != nil {
		log.Fatalf("Connect db error: %v", err)
	}

	if err := db.Ping(ctx); err != nil {
		log.Fatalf("Ping db error: %v", err)
	}

	//temporary
	//move all to init.sql
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
		log.Fatalf("Create table error: %v", err)
	}
	_, err = db.Exec(ctx,
		`INSERT INTO users(username, password, iin, role, registerDate) VALUES ($1, $2, $3, $4, $5)`,
		"admin", "pass", "940217200216", "admin", time.Now().Format("2006-01-02 15:04:05"))
	if err != nil {
		log.Printf("Admin already exist: %v", err)
	}
	return db
}
