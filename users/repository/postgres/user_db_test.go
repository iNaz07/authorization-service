package postgres_test

import (
	"testing"

	"github.com/driftprogramming/pgxpoolmock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	// "context"
	// "github.com/pashagolub/pgxmock"
	// "testing"
	// "time"
	// "transaction-service/domain"
	repo "transaction-service/users/repository/postgres"
)

func TestCreateUser(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// given
	mockPool := pgxpoolmock.NewMockPgxPool(ctrl)
	columns := []string{"iin", "username", "password", "role", "registerdate"}
	pgxRows := pgxpoolmock.NewRows(columns).AddRow("940217450216", "nazerke", "qwerty", "admin", "").ToPgxRows()
	mockPool.EXPECT().Query(gomock.Any(), gomock.Any(), gomock.Any()).Return(pgxRows, nil)
	userRepo := repo.NewUserRepository(mockPool)

	// when
	actualReq, err := userRepo.GetUserByID(1)

	// then
	assert.NotNil(t, actualReq)
	assert.NoError(t, err)
	assert.Equal(t, "940217450216", actualReq.IIN)
	assert.Equal(t, "nazerke", actualReq.Username)

}
