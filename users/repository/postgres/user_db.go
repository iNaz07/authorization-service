package postgres

import (
	"context"
	"fmt"
	"time"
	"transaction-service/domain"

	"github.com/jackc/pgx/v4/pgxpool"
)

type userRepository struct {
	Conn *pgxpool.Pool
}

func NewUserRepository(Conn *pgxpool.Pool) domain.UserRepository {
	return &userRepository{Conn}
}

func (u *userRepository) CreateUser(user *domain.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
	defer cancel()
	if _, err := u.Conn.Exec(ctx, "INSERT INTO users(iin, username, password, role, registerdate) VALUES ($1, $2, $3, $4, $5)",
		user.IIN, user.Username, user.Password, user.Role, user.RegisterDate); err != nil {
		return fmt.Errorf("dbInsertUser: %w", err)
	}
	return nil
}

func (u *userRepository) GetUserByID(id int64) (*domain.User, error) {

	user := &domain.User{}

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
	defer cancel()

	if err := u.Conn.QueryRow(ctx, "SELECT iin, username, role, registerdate FROM users WHERE id=$1", id).
		Scan(&user.IIN, &user.Username, &user.Role, &user.RegisterDate); err != nil {
		return nil, err
	}

	return user, nil
}

func (u *userRepository) GetUserByIIN(iin string) (*domain.User, error) {

	user := &domain.User{}

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
	defer cancel()

	if err := u.Conn.QueryRow(ctx, "SELECT id, iin, username, password FROM users WHERE iin=$1", iin).
		Scan(&user.ID, &user.IIN, &user.Username, &user.Password); err != nil {
		return nil, err
	}

	return user, nil
}

func (u *userRepository) GetUserByUsername(username string) (*domain.User, error) {

	user := &domain.User{}

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
	defer cancel()

	if err := u.Conn.QueryRow(ctx, "SELECT id, iin, username, password, role, registerDate FROM users WHERE username=$1", username).
		Scan(&user.ID, &user.IIN, &user.Username, &user.Password, &user.Role, &user.RegisterDate); err != nil {
		return nil, err
	}
	return user, nil
}

func (u *userRepository) GetAllUsers() ([]domain.User, error) {

	user := domain.User{}
	users := []domain.User{}

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
	defer cancel()

	rows, err := u.Conn.Query(ctx, "SELECT id, iin, username, role, registerdate FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&user.ID, &user.IIN, &user.Username, &user.Role, &user.RegisterDate); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}

func (u *userRepository) UpgradeUserRepo(username string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
	defer cancel()
	if _, err := u.Conn.Exec(ctx, "UPDATE users SET role=$1 WHERE username=$2",
		"admin", username); err != nil {
		return fmt.Errorf("db upgrade: %w", err)
	}
	return nil
}
