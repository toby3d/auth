package sqlite3_test

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	"source.toby3d.me/toby3d/auth/internal/domain"
	repository "source.toby3d.me/toby3d/auth/internal/session/repository/sqlite3"
	"source.toby3d.me/toby3d/auth/internal/testing/sqltest"
)

// nolint: gochecknoglobals // slices cannot be contants
var tableColumns = []string{"created_at", "code", "data"}

func TestCreate(t *testing.T) {
	t.Parallel()

	session := domain.TestSession(t)
	session.Profile = nil

	model, err := repository.NewSession(session)
	if err != nil {
		t.Fatal(err)
	}

	db, mock, cleanup := sqltest.Open(t)
	t.Cleanup(cleanup)

	createTable(t, mock)
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO sessions`)).
		WithArgs(
			sqltest.Time{},
			model.Code,
			model.Data,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	if err := repository.NewSQLite3SessionRepository(db).
		Create(context.Background(), *session); err != nil {
		t.Error(err)
	}
}

func TestGet(t *testing.T) {
	t.Parallel()

	session := domain.TestSession(t)
	session.Profile = nil

	model, err := repository.NewSession(session)
	if err != nil {
		t.Fatal(err)
	}

	db, mock, cleanup := sqltest.Open(t)
	t.Cleanup(cleanup)

	createTable(t, mock)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM sessions`)).
		WithArgs(session.Code).
		WillReturnRows(sqlmock.NewRows(tableColumns).
			AddRow(
				model.CreatedAt.Time,
				model.Code,
				model.Data,
			))

	result, err := repository.NewSQLite3SessionRepository(db).
		Get(context.Background(), session.Code)
	if err != nil {
		t.Fatal(err)
	}

	if result.Code != session.Code {
		t.Errorf("Get(%s) = %+v, want %+v", session.Code, result, session)
	}
}

func TestGetAndDelete(t *testing.T) {
	t.Parallel()

	session := domain.TestSession(t)
	session.Profile = nil

	model, err := repository.NewSession(session)
	if err != nil {
		t.Fatal(err)
	}

	db, mock, cleanup := sqltest.Open(t)
	t.Cleanup(cleanup)

	createTable(t, mock)
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM sessions`)).
		WithArgs(session.Code).
		WillReturnRows(sqlmock.NewRows(tableColumns).
			AddRow(
				model.CreatedAt.Time,
				model.Code,
				model.Data,
			))
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM sessions`)).
		WithArgs(model.Code).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	result, err := repository.NewSQLite3SessionRepository(db).
		GetAndDelete(context.Background(), session.Code)
	if err != nil {
		t.Fatal(err)
	}

	if result.Code != session.Code {
		t.Errorf("GetAndDelete(%s) = %+v, want %+v", session.Code, result, session)
	}
}

func createTable(tb testing.TB, mock sqlmock.Sqlmock) {
	tb.Helper()

	mock.ExpectExec(regexp.QuoteMeta(repository.QueryTable)).
		WillReturnResult(sqlmock.NewResult(1, 1))
}
