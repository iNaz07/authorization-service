package domain

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
	CreateUser(user *User) error
	GetUserByID(id int64) (*User, error)
	GetUserByUsername(username string) (*User, error)
	GetUserByIIN(iin string) (*User, error)
	GetAllUsers() ([]User, error)
	UpgradeUserRepo(username string) error
}

type UserUsecase interface {
	CreateUserUsecase(user *User) error
	GetUserByNameUsecase(name string) (*User, error) //пересмотреть
	GetUserByIINUsecase(iin string) (*User, error)
	GetUserByIDUsecase(id int64) (*User, error)
	GetAllUsecase() ([]User, error)
	UpgradeUserUsecase(username string) error
}
