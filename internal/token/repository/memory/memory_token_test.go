package memory_test

import (
	"context"
	"path"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"source.toby3d.me/website/indieauth/internal/domain"
	"source.toby3d.me/website/indieauth/internal/token"
	repository "source.toby3d.me/website/indieauth/internal/token/repository/memory"
)

func TestCreate(t *testing.T) {
	t.Parallel()

	store := new(sync.Map)
	accessToken := domain.TestToken(t)

	repo := repository.NewMemoryTokenRepository(store)
	require.NoError(t, repo.Create(context.TODO(), accessToken))

	result, ok := store.Load(path.Join(repository.DefaultPathPrefix, accessToken.AccessToken))
	assert.True(t, ok)
	assert.Equal(t, accessToken, result)

	assert.ErrorIs(t, repo.Create(context.TODO(), accessToken), token.ErrExist)
}

func TestGet(t *testing.T) {
	t.Parallel()

	store := new(sync.Map)
	accessToken := domain.TestToken(t)

	store.Store(path.Join(repository.DefaultPathPrefix, accessToken.AccessToken), accessToken)

	result, err := repository.NewMemoryTokenRepository(store).Get(context.TODO(), accessToken.AccessToken)
	assert.NoError(t, err)
	assert.Equal(t, accessToken, result)
}
