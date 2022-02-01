package memory_test

import (
	"context"
	"path"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"source.toby3d.me/website/indieauth/internal/domain"
	repository "source.toby3d.me/website/indieauth/internal/ticket/repository/memory"
)

func TestCreate(t *testing.T) {
	t.Parallel()

	store := new(sync.Map)
	ticket := domain.TestTicket(t)

	require.NoError(t, repository.NewMemoryTicketRepository(store, domain.TestConfig(t)).
		Create(context.TODO(), ticket))

	src, ok := store.Load(path.Join(repository.DefaultPathPrefix, ticket.Ticket))
	require.True(t, ok)

	result, ok := src.(*repository.Ticket)
	require.True(t, ok)
	assert.Equal(t, ticket, result.Ticket)
}

func TestGetAndDelete(t *testing.T) {
	t.Parallel()

	ticket := domain.TestTicket(t)

	store := new(sync.Map)
	store.Store(path.Join(repository.DefaultPathPrefix, ticket.Ticket), &repository.Ticket{
		CreatedAt: time.Now().UTC(),
		Ticket:    ticket,
	})

	result, err := repository.NewMemoryTicketRepository(store, domain.TestConfig(t)).
		GetAndDelete(context.TODO(), ticket.Ticket)
	require.NoError(t, err)
	assert.Equal(t, ticket, result)

	src, ok := store.Load(path.Join(repository.DefaultPathPrefix, ticket.Ticket))
	assert.False(t, ok)
	assert.Nil(t, src)
}
