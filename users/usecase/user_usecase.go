package usecase

import (
	"context"
	"fmt"
	"time"
	"transaction-service/domain"
	utils "transaction-service/utils"
)

type userUsecase struct {
	userRepo       domain.UserRepository
	timeoutContext time.Duration
}

func NewUserUseCase(repo domain.UserRepository, time time.Duration) domain.UserUsecase {
	return &userUsecase{userRepo: repo, timeoutContext: time}
}

func (u *userUsecase) CreateUserUsecase(ctx context.Context, user *domain.User) error {
	context, cancel := context.WithTimeout(ctx, u.timeoutContext)
	defer cancel()
	if _, err := u.userRepo.GetUserByIIN(context, user.IIN); err == nil {
		return fmt.Errorf("user already registered by iin: %w", err)
	}
	if err := utils.ValidateCreds(user.Username, user.Password, user.IIN); err != nil {
		return fmt.Errorf("invalid creds: %w", err)
	}
	hashedPassword := utils.GenerateHash(user.Password)
	user.Password = hashedPassword

	if user.Role == "" {
		user.Role = "user"
	}
	user.RegisterDate = time.Now().Format("2006-01-02 15:04:05")

	if err := u.userRepo.CreateUser(context, user); err != nil {
		return fmt.Errorf("registration error: %w", err)
	}
	return nil
}

func (u *userUsecase) GetUserByIDUsecase(ctx context.Context, id int64) (*domain.User, error) {
	context, cancel := context.WithTimeout(ctx, u.timeoutContext)
	defer cancel()

	user, err := u.userRepo.GetUserByID(context, id)
	if err != nil {
		return nil, fmt.Errorf("user doesn't exist: %w", err)
	}
	return user, nil
}

func (u *userUsecase) GetUserByNameUsecase(ctx context.Context, name string) (*domain.User, error) {
	context, cancel := context.WithTimeout(ctx, u.timeoutContext)
	defer cancel()

	user, err := u.userRepo.GetUserByUsername(context, name)
	if err != nil {
		return nil, fmt.Errorf("user doesn't exist: %w", err)
	}
	return user, nil
}

func (u *userUsecase) GetUserByIINUsecase(ctx context.Context, iin string) (*domain.User, error) {
	context, cancel := context.WithTimeout(ctx, u.timeoutContext)
	defer cancel()

	user, err := u.userRepo.GetUserByIIN(context, iin)
	if err != nil {
		return nil, fmt.Errorf("user doesn't exist: %w", err)
	}
	return user, nil
}

func (u *userUsecase) GetAllUsecase(ctx context.Context) ([]domain.User, error) {
	context, cancel := context.WithTimeout(ctx, u.timeoutContext)
	defer cancel()

	users, err := u.userRepo.GetAllUsers(context)
	if err != nil {
		return nil, fmt.Errorf("get all user error: %w", err)
	}
	return users, nil
}

func (u *userUsecase) UpgradeUserUsecase(ctx context.Context, username string) error {
	context, cancel := context.WithTimeout(ctx, u.timeoutContext)
	defer cancel()

	if err := u.userRepo.UpgradeUserRepo(context, username); err != nil {
		return fmt.Errorf("upgrade error: %w", err)
	}
	return nil
}
