package memory_test

import (
	"context"
	"path"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"source.toby3d.me/website/oauth/internal/domain"
	"source.toby3d.me/website/oauth/internal/token"
	repository "source.toby3d.me/website/oauth/internal/token/repository/memory"
)

func TestCreate(t *testing.T) {
	t.Parallel()

	store := new(sync.Map)
	repo := repository.NewMemoryTokenRepository(store)

	accessToken := domain.TestToken(t)
	require.NoError(t, repo.Create(context.TODO(), accessToken))

	expiry, ok := store.Load(path.Join(repository.DefaultPathPrefix, accessToken.AccessToken))
	assert.True(t, ok)
	assert.Equal(t, accessToken.Expiry, expiry)

	assert.EqualError(t, repo.Create(context.TODO(), accessToken), token.ErrExist.Error())
}

func TestGet(t *testing.T) {
	t.Parallel()

	store := new(sync.Map)
	repo := repository.NewMemoryTokenRepository(store)

	accessToken := domain.TestToken(t)
	store.Store(path.Join(repository.DefaultPathPrefix, accessToken.AccessToken), accessToken.Expiry)

	result, err := repo.Get(context.TODO(), accessToken.AccessToken)
	assert.NoError(t, err)
	assert.Equal(t, accessToken, result)
}
