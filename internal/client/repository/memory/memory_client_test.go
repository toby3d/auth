package memory_test

import (
	"context"
	"path"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	repository "source.toby3d.me/website/oauth/internal/client/repository/memory"
	"source.toby3d.me/website/oauth/internal/domain"
)

func TestGet(t *testing.T) {
	t.Parallel()

	store := new(sync.Map)
	client := domain.TestClient(t)

	store.Store(path.Join(repository.Key, client.ID), client)

	result, err := repository.NewMemoryClientRepository(store).Get(context.TODO(), client.ID)
	require.NoError(t, err)
	assert.Equal(t, client, result)
}
