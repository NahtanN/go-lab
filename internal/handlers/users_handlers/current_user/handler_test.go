package current_user

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/pashagolub/pgxmock/v3"
	"github.com/stretchr/testify/assert"

	"github.com/nahtann/go-lab/internal/context_values"
	"github.com/nahtann/go-lab/internal/storage/database/models"
)

func TestCurrentUser(t *testing.T) {
	user := models.UserModel{
		Id:        1,
		Username:  "Test User",
		Email:     "test@test.com",
		CreatedAt: time.Now(),
	}

	mock, err := pgxmock.NewPool(
		pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual),
	)
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mock.Close()

	rows := mock.NewRows([]string{"username", "email", "created_at"}).
		AddRow(user.Username, user.Email, user.CreatedAt)

	mock.ExpectQuery("SELECT username, email, created_at FROM users WHERE id = $1").
		WithArgs(user.Id).
		WillReturnRows(rows)

	handler := Handler{
		DB: mock,
	}

	request := Request{
		ID: 1,
	}

	response, err := handler.Exec(&request)

	assert.Nil(t, err)
	assert.NotNil(t, response)

	assert.Equal(t, user.Id, response.Id)
	assert.Equal(t, user.Username, response.Username)
	assert.Equal(t, user.Email, response.Email)
	assert.WithinDuration(t, user.CreatedAt, time.Now(), time.Second)
}

func TestCurrentUserFailOnDbError(t *testing.T) {
	mock, err := pgxmock.NewPool(
		pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual),
	)
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mock.Close()

	mock.ExpectQuery("SELECT username, email, created_at FROM users WHERE id = $1").
		WillReturnError(fmt.Errorf("Some error"))

	handler := Handler{
		DB: mock,
	}

	request := Request{
		ID: 1,
	}

	response, err := handler.Exec(&request)

	assert.Nil(t, response)
	assert.NotNil(t, err)

	assert.Equal(t, "Unable to retrieve current user data.", err.Error())
}

func TestCurrentUserParseDbData(t *testing.T) {
	user := models.UserModel{
		Id:       1,
		Username: "Test User",
	}

	mock, err := pgxmock.NewPool(
		pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual),
	)
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mock.Close()

	rows := mock.NewRows([]string{"1"}).
		AddRow(user.Username)

	mock.ExpectQuery("SELECT username, email, created_at FROM users WHERE id = $1").
		WithArgs(user.Id).
		WillReturnRows(rows)

	handler := Handler{
		DB: mock,
	}

	request := Request{
		ID: 1,
	}

	response, err := handler.Exec(&request)

	assert.Nil(t, response)
	assert.NotNil(t, err)

	assert.Equal(t, "Unable to parse current user data.", err.Error())
}

func TestRequestParser(t *testing.T) {
	ctx := context.WithValue(
		context.Background(),
		context_values.UserIdKey,
		float64(123),
	)

	request := Request{
		ID: 1,
	}

	req, err := http.NewRequest("GET", "/current", bytes.NewBufferString(""))
	if err != nil {
		t.Fatalf("an error '%s' was not expected while creating request", err)
	}

	req = req.WithContext(ctx)

	err = RequestParser(&request, req)

	assert.NoError(t, err)

	assert.Equal(t, uint32(123), request.ID)
}

func TestRequestParserError(t *testing.T) {
	ctx := context.WithValue(
		context.Background(),
		context_values.UserIdKey,
		nil,
	)

	request := Request{}

	req, err := http.NewRequest("GET", "/current", bytes.NewBufferString(""))
	if err != nil {
		t.Fatalf("an error '%s' was not expected while creating request", err)
	}

	req = req.WithContext(ctx)

	err = RequestParser(&request, req)

	assert.NotNil(t, err)
	assert.Equal(t, "Unable to fetch current user data.", err.Error())
}
