package http_test

import (
	"github.com/bxcodec/faker"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"transaction-service/domain"
	"transaction-service/domain/mocks"
	userHTTP "transaction-service/users/delivery/http"
	utils "transaction-service/utils"
)

func TestHome(t *testing.T) {

	var mockNewUser domain.User
	err := faker.FakeData(&mockNewUser)
	assert.NoError(t, err)

	mockUCase := new(mocks.UserUsecase)
	mockUCase.On("GetUserByIDUsecase", mock.Anything, mockNewUser.ID).Return(&mockNewUser, nil)

	e := echo.New()

	req, err := http.NewRequest(echo.GET, "/user/home", strings.NewReader(""))
	assert.NoError(t, err)
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)

	c.Set("user", mockNewUser)

	handler := userHTTP.UserHandler{
		UserUsecase: mockUCase,
	}
	err = handler.Home(c)
	require.NoError(t, err)

	mockUCase.AssertExpectations(t)
}

func TestRegistration(t *testing.T) {
	mockUser := &domain.User{
		Username: "nazerke",
		IIN:      "940217450216",
		Password: "Qwe12@",
	}

	tempMockUser := mockUser
	tempMockUser.ID = 0
	mockUCase := new(mocks.UserUsecase)

	mockUCase.On("CreateUserUsecase", mock.Anything, mockUser).Return(nil)

	e := echo.New()
	req, err := http.NewRequest(echo.POST, "/signup?username=nazerke&iin=940217450216&password=Qwe12@", strings.NewReader(""))
	assert.NoError(t, err)

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/signup")

	handler := userHTTP.UserHandler{
		UserUsecase: mockUCase,
	}
	err = handler.Registration(c)
	require.NoError(t, err)

	assert.Equal(t, http.StatusCreated, rec.Code)
	mockUCase.AssertExpectations(t)
}

func TestSignIn(t *testing.T) {
	var mockNewUser *domain.User
	err := faker.FakeData(&mockNewUser)
	assert.NoError(t, err)
	psw := mockNewUser.Password
	var token string

	mockUCase := new(mocks.UserUsecase)
	mockUCase.On("GetUserByNameUsecase", mock.Anything, mockNewUser.Username).Return(mockNewUser, nil)
	ok := utils.ComparePasswordHash(mockNewUser.Password, psw)
	assert.False(t, ok)
	mockJWTUCase := new(mocks.JwtTokenUsecase)
	mockJWTUCase.On("GenerateToken", mockNewUser.ID, mockNewUser.Role, mockNewUser.IIN).Return(token, nil)
	mockJWTUCase.On("InsertToken", mockNewUser.ID, token).Return(nil)

	e := echo.New()
	req, err := http.NewRequest(echo.POST, "/sigin?username="+mockNewUser.Username+"&password="+mockNewUser.Password, strings.NewReader(""))
	assert.NoError(t, err)

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/signin")

	handler := userHTTP.UserHandler{
		UserUsecase: mockUCase,
		JwtUsecase:  mockJWTUCase,
	}
	err = handler.Signin(c)
	require.NoError(t, err)

	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockUCase.AssertExpectations(t)
}

func TestGetAllUserInfo(t *testing.T) {

	var mockNewUser domain.User
	err := faker.FakeData(&mockNewUser)
	assert.NoError(t, err)
	mockNewUser.Role = "admin"
	mockListUser := make([]domain.User, 0)
	mockListUser = append(mockListUser, mockNewUser)

	mockUCase := new(mocks.UserUsecase)
	mockUCase.On("GetAllUsecase", mock.Anything).Return(mockListUser, nil)

	e := echo.New()
	req, err := http.NewRequest(echo.GET, "/user/info/all", strings.NewReader(""))
	assert.NoError(t, err)

	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)

	c.Set("user", mockNewUser)
	c.SetPath("/user/info/all")

	handler := userHTTP.UserHandler{
		UserUsecase: mockUCase,
	}
	err = handler.GetAllUserInfo(c)
	require.NoError(t, err)

	mockUCase.AssertExpectations(t)
}

func TestGetUserInfo(t *testing.T) {

	var mockNewUser domain.User
	err := faker.FakeData(&mockNewUser)
	assert.NoError(t, err)
	id := strconv.Itoa(int(mockNewUser.ID))
	mockUCase := new(mocks.UserUsecase)
	mockUCase.On("GetUserByIDUsecase", mock.Anything, mockNewUser.ID).Return(&mockNewUser, nil)

	e := echo.New()
	req, err := http.NewRequest(echo.GET, "/user/info/"+id, strings.NewReader(""))
	assert.NoError(t, err)

	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)

	c.Set("user", mockNewUser)
	c.SetPath("/user/info/:id")
	c.SetParamNames("id")
	c.SetParamValues(id)

	handler := userHTTP.UserHandler{
		UserUsecase: mockUCase,
	}
	err = handler.GetUserInfo(c)
	require.NoError(t, err)

	mockUCase.AssertExpectations(t)
}

func TestUpgradeRole(t *testing.T) {

	var mockNewUser domain.User
	err := faker.FakeData(&mockNewUser)
	assert.NoError(t, err)
	mockNewUser.Role = "admin"

	mockUCase := new(mocks.UserUsecase)
	mockUCase.On("UpgradeUserUsecase", mock.Anything, "someuser").Return(nil)

	e := echo.New()

	req, err := http.NewRequest(echo.GET, "/user/info/someuser", strings.NewReader(""))
	assert.NoError(t, err)
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)

	c.Set("user", mockNewUser)
	c.SetPath("/user/info/:username")
	c.SetParamNames("username")
	c.SetParamValues("someuser")

	handler := userHTTP.UserHandler{
		UserUsecase: mockUCase,
	}
	err = handler.UpgradeRole(c)
	require.NoError(t, err)

	mockUCase.AssertExpectations(t)
}
