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
	if _, err := u.Conn.Exec(ctx, "INSERT INTO users(iin, username, password, role) VALUES ($1, $2, $3, $4)",
		user.IIN, user.Username, user.Password, user.Role); err != nil {
		return fmt.Errorf("dbInsertUser: %w", err)
	}
	return nil
}

func (u *userRepository) AddRoleToUser(id int, role string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
	defer cancel()
	if _, err := u.Conn.Exec(ctx, "UPDATE users SET role=$1 WHERE id=$2",
		role, id); err != nil {
		return fmt.Errorf("dbInsertUser: %w", err)
	}
	return nil
}

func (u *userRepository) GetUserByID(id int64) (*domain.User, error) {

	user := &domain.User{}

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
	defer cancel()

	if err := u.Conn.QueryRow(ctx, "SELECT id, iin, username, password FROM users WHERE id=$1", id).
		Scan(&user.ID, &user.IIN, &user.Username, &user.Password); err != nil {
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

	if err := u.Conn.QueryRow(ctx, "SELECT id, iin, username, password FROM users WHERE username=$1", username).
		Scan(&user.ID, &user.IIN, &user.Username, &user.Password); err != nil {
		return nil, err
	}
	return user, nil
}

func (u *userRepository) GetAllUsers() ([]*domain.User, error) {

	user := &domain.User{}
	users := []*domain.User{}

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
	defer cancel()

	rows, err := u.Conn.Query(ctx, "SELECT id, iin, username, password FROM users")
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		if err := rows.Scan(&user.ID, &user.IIN, &user.Username, &user.Password); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}

func (u *userRepository) DeleteUserByIIN(iin string) error {

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
	defer cancel()

	_, err := u.Conn.Exec(ctx, "DELETE FROM users WHERE iin=$1", iin)
	return err
}
