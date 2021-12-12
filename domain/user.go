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
	User            User
	Number          string `json:"number"`
	Balance         int64  `json:"balance"`
	RegisterDate    string `json:"registerDate"`
	LastTransaction string `json:"lasttransaction"`
}

type UserRepository interface {
	CreateUser(user *User) error
	GetUserByID(id int64) (*User, error)              //надо подумать нужно ли?
	GetUserByUsername(username string) (*User, error) // то же, можно ли допускать одинаковые имена, или это уникальные ники?
	GetUserByIIN(iin string) (*User, error)
	GetAllUsers() ([]User, error)
	DeleteUserByIIN(iin string) error //зависит от предыдущих ответов
}

type UserUsecase interface {
	CreateUserUsecase(user *User) error
	GetUserByNameUsecase(name string) (*User, error) //пересмотреть
	GetUserByIINUsecase(iin string) (*User, error)
	GetUserByIDUsecase(id int64) (*User, error)
	GetAllUsecase() ([]User, error)
	DeleteUserUsecase(iin string) error
}
