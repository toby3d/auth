package memory_test

import (
	"context"
	"path"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"source.toby3d.me/website/oauth/internal/domain"
	repository "source.toby3d.me/website/oauth/internal/token/repository/memory"
)

func TestCreate(t *testing.T) {
	t.Parallel()

	store := new(sync.Map)
	token := domain.TestToken(t)

	repo := repository.NewMemoryTokenRepository(store)
	require.NoError(t, repo.Create(context.TODO(), token))

	result, ok := store.Load(path.Join(repository.DefaultPathPrefix, token.AccessToken))
	assert.True(t, ok)
	assert.Equal(t, token, result)

	assert.ErrorIs(t, repo.Create(context.TODO(), token), repository.ErrExist)
}

func TestGet(t *testing.T) {
	t.Parallel()

	store := new(sync.Map)
	token := domain.TestToken(t)

	store.Store(path.Join(repository.DefaultPathPrefix, token.AccessToken), token)

	result, err := repository.NewMemoryTokenRepository(store).Get(context.TODO(), token.AccessToken)
	assert.NoError(t, err)
	assert.Equal(t, token, result)
}
