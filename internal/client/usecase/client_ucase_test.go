package usecase_test

import (
	"context"
	"path"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	repository "source.toby3d.me/website/oauth/internal/client/repository/memory"
	"source.toby3d.me/website/oauth/internal/client/usecase"
	"source.toby3d.me/website/oauth/internal/domain"
)

func TestDiscovery(t *testing.T) {
	t.Parallel()

	client := domain.TestClient(t)

	store := new(sync.Map)
	store.Store(path.Join(repository.DefaultPathPrefix, client.ID.String()), client)

	result, err := usecase.NewClientUseCase(repository.NewMemoryClientRepository(store)).
		Discovery(context.TODO(), client.ID)
	require.NoError(t, err)
	assert.Equal(t, client, result)
}
