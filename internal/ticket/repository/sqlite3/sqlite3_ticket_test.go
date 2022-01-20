package sqlite3_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"source.toby3d.me/website/indieauth/internal/domain"
	"source.toby3d.me/website/indieauth/internal/testing/sqltest"
	repository "source.toby3d.me/website/indieauth/internal/ticket/repository/sqlite3"
)

func TestCreate(t *testing.T) {
	t.Parallel()

	db, cleanup := sqltest.Open(t)
	t.Cleanup(cleanup)

	ticket := domain.TestTicket(t)
	require.NoError(t, repository.NewSQLite3TicketRepository(db, domain.TestConfig(t)).
		Create(context.Background(), ticket))

	results := make([]*repository.Ticket, 0)
	require.NoError(t, db.Select(&results, "SELECT * FROM tickets;"))
	require.Len(t, results, 1)

	result := new(domain.Ticket)
	results[0].Populate(result)

	assert.Equal(t, ticket.Ticket, result.Ticket)
}

func TestGetAndDelete(t *testing.T) {
	t.Parallel()

	db, cleanup := sqltest.Open(t)
	t.Cleanup(cleanup)

	ticket := domain.TestTicket(t)
	_, err := db.NamedExec(repository.QueryTable+repository.QueryCreate, repository.NewTicket(ticket))
	require.NoError(t, err)

	result, err := repository.NewSQLite3TicketRepository(db, domain.TestConfig(t)).
		GetAndDelete(context.Background(), ticket.Ticket)
	require.NoError(t, err)
	assert.Equal(t, ticket.Ticket, result.Ticket)

	results := make([]*repository.Ticket, 0)
	require.NoError(t, db.Select(&results, "SELECT * FROM tickets;"))
	assert.Empty(t, results)
}
