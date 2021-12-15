package postgres_test

import (
	"context"
	"github.com/pashagolub/pgxmock"
	"testing"
	"time"
	"transaction-service/domain"
	repo "transaction-service/users/repository/postgres"
)

func TestCreateUser(t *testing.T) {
	mock, err := pgxmock.NewConn()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mock.Close(context.Background())
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO users\\(iin, username, password, role, registerdate\\)").
		WithArgs("940217450216", "nazerke", "Ekrezan!94", "admin", time.Now().Format("2006-01-02 15:04:05")).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))
	mock.ExpectCommit()
	conn := repo.NewUserRepository(mock.Conn())
	if err = conn.CreateUser(&domain.User{
		IIN:          "940217450216",
		Username:     "nazerke",
		Password:     "Ekrezan!94",
		Role:         "admin",
		RegisterDate: time.Now().Format("2006-01-02 15:04:05"),
	}); err != nil {

	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}

}
