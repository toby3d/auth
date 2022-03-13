package sqlite3_test

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	"source.toby3d.me/toby3d/auth/internal/domain"
	"source.toby3d.me/toby3d/auth/internal/testing/sqltest"
	repository "source.toby3d.me/toby3d/auth/internal/token/repository/sqlite3"
)

//nolint: gochecknoglobals // slices cannot be contants
var tableColumns = []string{"created_at", "access_token", "client_id", "me", "scope"}

func TestCreate(t *testing.T) {
	t.Parallel()

	token := domain.TestToken(t)
	model := repository.NewToken(token)

	db, mock, cleanup := sqltest.Open(t)
	t.Cleanup(cleanup)

	createTable(t, mock)
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO tokens`)).
		WithArgs(
			sqltest.Time{},
			model.AccessToken,
			model.ClientID,
			model.Me,
			model.Scope,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	if err := repository.NewSQLite3TokenRepository(db).Create(context.TODO(), token); err != nil {
		t.Error(err)
	}
}

func TestGet(t *testing.T) {
	t.Parallel()

	token := domain.TestToken(t)
	model := repository.NewToken(token)

	db, mock, cleanup := sqltest.Open(t)
	t.Cleanup(cleanup)

	createTable(t, mock)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM tokens`)).
		WithArgs(model.AccessToken).
		WillReturnRows(sqlmock.NewRows(tableColumns).
			AddRow(
				model.CreatedAt.Time,
				model.AccessToken,
				model.ClientID,
				model.Me,
				model.Scope,
			))

	result, err := repository.NewSQLite3TokenRepository(db).Get(context.TODO(), token.AccessToken)
	if err != nil {
		t.Fatal(err)
	}

	if result.AccessToken != token.AccessToken {
		t.Errorf("Get(%s) = %+v, want %+v", token.AccessToken, result, token)
	}
}

func createTable(tb testing.TB, mock sqlmock.Sqlmock) {
	tb.Helper()

	mock.ExpectExec(regexp.QuoteMeta(repository.QueryTable)).
		WillReturnResult(sqlmock.NewResult(1, 1))
}
