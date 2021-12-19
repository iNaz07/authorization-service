package domain

import (
	"context"
)

type User struct {
	ID           int64  `json:"id"`
	IIN          string `json:"iin"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	Role         string `json:"role"`
	RegisterDate string `json:"registerdate"`
}

type Accounts struct {
	Number          string `json:"number"`
	Balance         int64  `json:"balance"`
	RegisterDate    string `json:"registerDate"`
	LastTransaction string `json:"lasttransaction"`
}

type UserInfo struct {
	User     User
	Accounts []Accounts
}

type UserRepository interface {
	CreateUser(ctx context.Context, user *User) error
	GetUserByID(ctx context.Context, id int64) (*User, error)
	GetUserByUsername(ctx context.Context, username string) (*User, error)
	GetUserByIIN(ctx context.Context, iin string) (*User, error)
	GetAllUsers(ctx context.Context) ([]User, error)
	UpgradeUserRepo(ctx context.Context, username string) error
}

type UserUsecase interface {
	CreateUserUsecase(ctx context.Context, user *User) error
	GetUserByNameUsecase(ctx context.Context, name string) (*User, error) //пересмотреть
	GetUserByIINUsecase(ctx context.Context, iin string) (*User, error)
	GetUserByIDUsecase(ctx context.Context, id int64) (*User, error)
	GetAllUsecase(ctx context.Context) ([]User, error)
	UpgradeUserUsecase(ctx context.Context, username string) error
}
