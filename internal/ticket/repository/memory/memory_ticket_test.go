package memory_test

import (
	"context"
	"path"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"source.toby3d.me/website/oauth/internal/domain"
	repository "source.toby3d.me/website/oauth/internal/ticket/repository/memory"
)

func TestGet(t *testing.T) {
	t.Parallel()

	ticket := domain.TestTicket(t)
	user := domain.TestUser(t)

	store := new(sync.Map)
	store.Store(path.Join(repository.DefaultPathPrefix, ticket.Resource.String()), user.TokenEndpoint)

	result, err := repository.NewMemoryTicketRepository(store).Get(context.Background(), ticket.Resource)
	require.NoError(t, err)
	assert.Equal(t, user.TokenEndpoint.String(), result.String())
}
