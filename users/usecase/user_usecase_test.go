package usecase_test

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"

	"transaction-service/domain"
	"transaction-service/domain/mocks"
	ucase "transaction-service/users/usecase"
	utils "transaction-service/utils"
)

func TestCreateUser(t *testing.T) {
	mockUserRepo := new(mocks.UserRepository)
	mockUser := &domain.User{
		Username: "jack",
		Password: "QWEqwe123!!@#",
		IIN:      "940217450216",
	}

	t.Run("success", func(t *testing.T) {
		mockUserRepo.On("GetUserByIIN", mock.Anything, mock.AnythingOfType("string")).Return(nil, errors.New("no rows in result set")).Once()
		err := utils.ValidateCreds(mockUser.Username, mockUser.Password, mockUser.IIN)
		assert.NoError(t, err)
		mockUserRepo.On("CreateUser", mock.Anything, mock.AnythingOfType("*domain.User")).Return(nil).Once()
		u := ucase.NewUserUseCase(mockUserRepo, 2*time.Second)
		err = u.CreateUserUsecase(context.Background(), mockUser)

		assert.NoError(t, err)

		mockUserRepo.AssertExpectations(t)
	})
	t.Run("error-failed", func(t *testing.T) {
		newMockUser := &domain.User{
			Username: "jack",
			Password: "qwerty",
			IIN:      "940217450216",
		}
		mockUserRepo.On("GetUserByIIN",mock.Anything, mock.AnythingOfType("string")).Return(nil, errors.New("no rows in result set")).Once()
		err := utils.ValidateCreds(newMockUser.Username, newMockUser.Password, newMockUser.IIN)
		assert.EqualError(t, err, "password must contain at least 1 digit, 1 uppercase and 1 lowercase letter")
	})
}

func TestGetUserByIDUsecase(t *testing.T) {
	mockUserRepo := new(mocks.UserRepository)
	mockUser := &domain.User{
		ID:           25,
		IIN:          "940217450216",
		Username:     "content",
		Password:     "Qwe123@",
		Role:         "user",
		RegisterDate: time.Now().Format("2006-01-02 15:04:05"),
	}

	t.Run("success", func(t *testing.T) {
		mockUserRepo.On("GetUserByID", mock.Anything, mock.AnythingOfType("int64")).Return(mockUser, nil).Once()

		u := ucase.NewUserUseCase(mockUserRepo, 2*time.Second)

		a, err := u.GetUserByIDUsecase(context.Background(), mockUser.ID)

		assert.NoError(t, err)
		assert.NotNil(t, a)

		mockUserRepo.AssertExpectations(t)
	})
	t.Run("error-failed", func(t *testing.T) {
		mockUserRepo.On("GetUserByID", mock.Anything, mock.AnythingOfType("int64")).Return(&domain.User{}, errors.New("Unexpected")).Once()

		u := ucase.NewUserUseCase(mockUserRepo, 2*time.Second)

		a, err := u.GetUserByIDUsecase(context.Background(), mockUser.ID)

		assert.Error(t, err)
		assert.NotSame(t, &domain.User{}, a)

		mockUserRepo.AssertExpectations(t)
	})
}

func TestGetUserByNameUsecase(t *testing.T) {
	mockUserRepo := new(mocks.UserRepository)
	mockUser := &domain.User{
		ID:           25,
		IIN:          "940217450216",
		Username:     "content",
		Password:     "Qwe123@",
		Role:         "user",
		RegisterDate: time.Now().Format("2006-01-02 15:04:05"),
	}

	t.Run("success", func(t *testing.T) {
		mockUserRepo.On("GetUserByUsername", mock.Anything, mock.AnythingOfType("string")).Return(mockUser, nil).Once()

		u := ucase.NewUserUseCase(mockUserRepo, 2*time.Second)

		a, err := u.GetUserByNameUsecase(context.Background(), mockUser.Username)

		assert.NoError(t, err)
		assert.NotNil(t, a)

		mockUserRepo.AssertExpectations(t)
	})
	t.Run("error-failed", func(t *testing.T) {
		mockUserRepo.On("GetUserByUsername", mock.Anything, mock.AnythingOfType("string")).Return(&domain.User{}, errors.New("Unexpected")).Once()

		u := ucase.NewUserUseCase(mockUserRepo, 2*time.Second)

		a, err := u.GetUserByNameUsecase(context.Background(), mockUser.Username)

		assert.Error(t, err)
		assert.NotSame(t, &domain.User{}, a)

		mockUserRepo.AssertExpectations(t)
	})
}

func TestGetAllUsecase(t *testing.T) {
	mockUserRepo := new(mocks.UserRepository)
	mockUser := []domain.User{
		{
			ID:           1,
			IIN:          "940217450216",
			Username:     "content",
			Password:     "Qwe123@",
			Role:         "user",
			RegisterDate: time.Now().Format("2006-01-02 15:04:05"),
		},
		{
			ID:           2,
			IIN:          "940217450016",
			Username:     "naz",
			Password:     "Qwe123@",
			Role:         "user",
			RegisterDate: time.Now().Format("2006-01-02 15:04:05"),
		},
	}

	t.Run("success", func(t *testing.T) {
		mockUserRepo.On("GetAllUsers", mock.Anything).Return(mockUser, nil).Once()

		u := ucase.NewUserUseCase(mockUserRepo, 2*time.Second)

		a, err := u.GetAllUsecase(context.Background())

		assert.NoError(t, err)
		assert.NotNil(t, a)

		mockUserRepo.AssertExpectations(t)
	})
	t.Run("error-failed", func(t *testing.T) {
		mockUserRepo.On("GetAllUsers", mock.Anything).Return([]domain.User{}, errors.New("Unexpected")).Once()

		u := ucase.NewUserUseCase(mockUserRepo, 2*time.Second)

		a, err := u.GetAllUsecase(context.Background())

		assert.Error(t, err)
		assert.NotSame(t, &domain.User{}, a)

		mockUserRepo.AssertExpectations(t)
	})
}

func TestUpgradeUserUsecase(t *testing.T) {
	mockUserRepo := new(mocks.UserRepository)
	username := "nazerke"

	t.Run("success", func(t *testing.T) {
		mockUserRepo.On("UpgradeUserRepo", mock.Anything, mock.AnythingOfType("string")).Return(nil).Once()

		u := ucase.NewUserUseCase(mockUserRepo, 2*time.Second)

		err := u.UpgradeUserUsecase(context.Background(), username)
		assert.NoError(t, err)

		mockUserRepo.AssertExpectations(t)
	})
}
