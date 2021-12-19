package usecase

import (
	"context"
	"net/http"
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
		return &domain.LogError{"user already registered by iin", err, http.StatusBadRequest}
	}
	if err := utils.ValidateCreds(user.Username, user.Password, user.IIN); err != nil {
		return &domain.LogError{"invalid credentials", err, http.StatusBadRequest}
	}
	// TODO: fix ths func
	hashedPassword := utils.GenerateHash(user.Password)
	user.Password = hashedPassword

	if user.Role == "" {
		user.Role = "user"
	}
	user.RegisterDate = time.Now().Format("2006-01-02 15:04:05")

	if err := u.userRepo.CreateUser(context, user); err != nil {
		return &domain.LogError{"registration error", err, http.StatusInternalServerError}
	}
	return nil
}

func (u *userUsecase) GetUserByIDUsecase(ctx context.Context, id int64) (*domain.User, error) {
	context, cancel := context.WithTimeout(ctx, u.timeoutContext)
	defer cancel()

	user, err := u.userRepo.GetUserByID(context, id)
	if err != nil {
		return nil, &domain.LogError{"user not found", err, http.StatusNotFound}
	}
	return user, nil
}

func (u *userUsecase) GetUserByNameUsecase(ctx context.Context, name string) (*domain.User, error) {
	context, cancel := context.WithTimeout(ctx, u.timeoutContext)
	defer cancel()

	user, err := u.userRepo.GetUserByUsername(context, name)
	if err != nil {
		return nil, &domain.LogError{"user not found", err, http.StatusNotFound}
	}
	return user, nil
}

func (u *userUsecase) GetUserByIINUsecase(ctx context.Context, iin string) (*domain.User, error) {
	context, cancel := context.WithTimeout(ctx, u.timeoutContext)
	defer cancel()

	user, err := u.userRepo.GetUserByIIN(context, iin)
	if err != nil {
		return nil, &domain.LogError{"user not found", err, http.StatusNotFound}
	}
	return user, nil
}

func (u *userUsecase) GetAllUsecase(ctx context.Context) ([]domain.User, error) {
	context, cancel := context.WithTimeout(ctx, u.timeoutContext)
	defer cancel()

	users, err := u.userRepo.GetAllUsers(context)
	if err != nil {
		return nil, &domain.LogError{"GetAllUser error", err, http.StatusInternalServerError}
	}
	return users, nil
}

func (u *userUsecase) UpgradeUserUsecase(ctx context.Context, username string) error {
	context, cancel := context.WithTimeout(ctx, u.timeoutContext)
	defer cancel()

	if _, err := u.userRepo.GetUserByUsername(context, username); err != nil {
		return &domain.LogError{"user not found to upgrade role", err, http.StatusBadRequest}
	}

	if err := u.userRepo.UpgradeUserRepo(context, username); err != nil {
		return &domain.LogError{"cannot upgrade user role", err, http.StatusInternalServerError}
	}
	return nil
}
