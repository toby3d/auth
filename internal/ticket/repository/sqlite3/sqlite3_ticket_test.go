package sqlite3_test

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	"source.toby3d.me/toby3d/auth/internal/domain"
	"source.toby3d.me/toby3d/auth/internal/testing/sqltest"
	repository "source.toby3d.me/toby3d/auth/internal/ticket/repository/sqlite3"
)

//nolint: gochecknoglobals // slices cannot be contants
var tableColumns = []string{"created_at", "resource", "subject", "ticket"}

func TestCreate(t *testing.T) {
	t.Parallel()

	ticket := domain.TestTicket(t)
	model := repository.NewTicket(ticket)
	db, mock, cleanup := sqltest.Open(t)
	t.Cleanup(cleanup)

	createTable(t, mock)
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO tickets`)).
		WithArgs(
			sqltest.Time{},
			model.Resource,
			model.Subject,
			model.Ticket,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	if err := repository.NewSQLite3TicketRepository(db, domain.TestConfig(t)).
		Create(context.Background(), *ticket); err != nil {
		t.Error(err)
	}
}

func TestGetAndDelete(t *testing.T) {
	t.Parallel()

	ticket := domain.TestTicket(t)
	model := repository.NewTicket(ticket)
	db, mock, cleanup := sqltest.Open(t)
	t.Cleanup(cleanup)

	createTable(t, mock)
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM tickets`)).
		WithArgs(model.Ticket).
		WillReturnRows(sqlmock.NewRows(tableColumns).
			AddRow(
				model.CreatedAt.Time,
				model.Resource,
				model.Subject,
				model.Ticket,
			))
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM tickets`)).
		WithArgs(model.Ticket).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	result, err := repository.NewSQLite3TicketRepository(db, domain.TestConfig(t)).
		GetAndDelete(context.Background(), ticket.Ticket)
	if err != nil {
		t.Fatal(err)
	}

	if result.Ticket != ticket.Ticket {
		t.Errorf("GetAndDelete(%s) = %+v, want %+v", ticket.Ticket, result, ticket)
	}
}

func createTable(tb testing.TB, mock sqlmock.Sqlmock) {
	tb.Helper()

	mock.ExpectExec(regexp.QuoteMeta(repository.QueryTable)).
		WillReturnResult(sqlmock.NewResult(1, 1))
}
