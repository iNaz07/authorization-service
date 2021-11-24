package usecase

import (
	"fmt"
	"transaction-service/domain"
)

type userUsecase struct {
	userRepo domain.UserRepository
}

func NewUserUseCase(repo domain.UserRepository) domain.UserUsecase {
	return &userUsecase{repo}
}

//TODO: нужно зашифровать пароль
//TODO: validate name, password, iin
func (u *userUsecase) CreateUserUsecase(user *domain.User) error {
	if _, err := u.userRepo.GetUserByIIN(user.IIN); err == nil {
		return fmt.Errorf("user already registered by iin: %w", err)
	}
	if err := u.userRepo.CreateUser(user); err != nil {
		return fmt.Errorf("registration error: %w", err)
	}
	return nil
}

func (u *userUsecase) GetUserByNameUsecase(name string) (*domain.User, error) {
	user, err := u.userRepo.GetUserByUsername(name)
	if err != nil {
		return nil, fmt.Errorf("user doesn't exist: %w", err)
	}
	return user, nil
}

func (u *userUsecase) GetUserByIINUsecase(iin string) (*domain.User, error) {
	user, err := u.userRepo.GetUserByIIN(iin)
	if err != nil {
		return nil, fmt.Errorf("user doesn't exist: %w", err)
	}
	return user, nil
}

func (u *userUsecase) GetAllUsecase() ([]*domain.User, error) {
	users, err := u.userRepo.GetAllUsers()
	if err != nil {
		return nil, fmt.Errorf("user doesn't exist: %w", err)
	}
	return users, nil
}

func (u *userUsecase) DeleteUserUsecase(iin string) error {
	if err := u.userRepo.DeleteUserByIIN(iin); err != nil {
		return fmt.Errorf("delete user error: %w", err)
	}
	return nil
}
