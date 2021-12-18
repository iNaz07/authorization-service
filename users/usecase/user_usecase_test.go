package usecase_test

import (
	// "context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

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
		mockUserRepo.On("GetUserByIIN", mock.AnythingOfType("string")).Return(nil, errors.New("no rows in result set")).Once()
		err := utils.ValidateCreds(mockUser.Username, mockUser.Password, mockUser.IIN)
		assert.NoError(t, err)
		mockUserRepo.On("CreateUser", mock.AnythingOfType("*domain.User")).Return(nil).Once()
		u := ucase.NewUserUseCase(mockUserRepo)
		err = u.CreateUserUsecase(mockUser)

		assert.NoError(t, err)

		mockUserRepo.AssertExpectations(t)
	})
	t.Run("error-failed", func(t *testing.T) {
		newMockUser := &domain.User{
			Username: "jack",
			Password: "qwerty",
			IIN:      "940217450216",
		}
		mockUserRepo.On("GetUserByIIN", mock.AnythingOfType("string")).Return(nil, errors.New("no rows in result set")).Once()
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
		mockUserRepo.On("GetUserByID", mock.AnythingOfType("int64")).Return(mockUser, nil).Once()

		u := ucase.NewUserUseCase(mockUserRepo)

		a, err := u.GetUserByIDUsecase(mockUser.ID)

		assert.NoError(t, err)
		assert.NotNil(t, a)

		mockUserRepo.AssertExpectations(t)
	})
	t.Run("error-failed", func(t *testing.T) {
		mockUserRepo.On("GetUserByID", mock.AnythingOfType("int64")).Return(&domain.User{}, errors.New("Unexpected")).Once()

		u := ucase.NewUserUseCase(mockUserRepo)

		a, err := u.GetUserByIDUsecase(mockUser.ID)

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
		mockUserRepo.On("GetUserByUsername", mock.AnythingOfType("string")).Return(mockUser, nil).Once()

		u := ucase.NewUserUseCase(mockUserRepo)

		a, err := u.GetUserByNameUsecase(mockUser.Username)

		assert.NoError(t, err)
		assert.NotNil(t, a)

		mockUserRepo.AssertExpectations(t)
	})
	t.Run("error-failed", func(t *testing.T) {
		mockUserRepo.On("GetUserByUsername", mock.AnythingOfType("string")).Return(&domain.User{}, errors.New("Unexpected")).Once()

		u := ucase.NewUserUseCase(mockUserRepo)

		a, err := u.GetUserByNameUsecase(mockUser.Username)

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
		mockUserRepo.On("GetAllUsers").Return(mockUser, nil).Once()

		u := ucase.NewUserUseCase(mockUserRepo)

		a, err := u.GetAllUsecase()

		assert.NoError(t, err)
		assert.NotNil(t, a)

		mockUserRepo.AssertExpectations(t)
	})
	t.Run("error-failed", func(t *testing.T) {
		mockUserRepo.On("GetAllUsers").Return([]domain.User{}, errors.New("Unexpected")).Once()

		u := ucase.NewUserUseCase(mockUserRepo)

		a, err := u.GetAllUsecase()

		assert.Error(t, err)
		assert.NotSame(t, &domain.User{}, a)

		mockUserRepo.AssertExpectations(t)
	})
}

func TestUpgradeUserUsecase(t *testing.T) {
	mockUserRepo := new(mocks.UserRepository)
	username := "nazerke"

	t.Run("success", func(t *testing.T) {
		mockUserRepo.On("UpgradeUserRepo", mock.AnythingOfType("string")).Return(nil).Once()

		u := ucase.NewUserUseCase(mockUserRepo)

		err := u.UpgradeUserUsecase(username)
		assert.NoError(t, err)

		mockUserRepo.AssertExpectations(t)
	})
}
