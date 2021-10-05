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

func TestGet(t *testing.T) {
	t.Parallel()

	store := new(sync.Map)
	accessToken := domain.TestToken(t)

	store.Store(path.Join(repository.Key, accessToken.AccessToken), accessToken)

	result, err := repository.NewMemoryTokenRepository(store).Get(context.TODO(), accessToken.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, accessToken, result)
}

func TestCreate(t *testing.T) {
	t.Parallel()

	store := new(sync.Map)
	accessToken := domain.TestToken(t)

	repo := repository.NewMemoryTokenRepository(store)
	require.NoError(t, repo.Create(context.TODO(), accessToken))

	result, ok := store.Load(path.Join(repository.Key, accessToken.AccessToken))
	assert.True(t, ok)
	assert.Equal(t, accessToken, result)

	assert.EqualError(t, repo.Create(context.TODO(), accessToken), token.ErrExist.Error())
}

func TestUpdate(t *testing.T) {
	t.Parallel()

	store := new(sync.Map)
	accessToken := domain.TestToken(t)

	store.Store(path.Join(repository.Key, accessToken.AccessToken), accessToken)

	tokenCopy := *accessToken
	tokenCopy.ClientID = "https://client.example.com/"
	tokenCopy.Me = "https://toby3d.ru/"

	require.NoError(t, repository.NewMemoryTokenRepository(store).Update(context.TODO(), &tokenCopy))

	result, ok := store.Load(path.Join(repository.Key, accessToken.AccessToken))
	assert.True(t, ok)
	assert.NotEqual(t, accessToken, result)
	assert.Equal(t, &tokenCopy, result)
}

func TestDelete(t *testing.T) {
	t.Parallel()

	store := new(sync.Map)
	accessToken := domain.TestToken(t)

	store.Store(path.Join(repository.Key, accessToken.AccessToken), accessToken)

	require.NoError(t, repository.NewMemoryTokenRepository(store).Remove(context.TODO(), accessToken.AccessToken))

	result, ok := store.Load(path.Join(repository.Key, accessToken.AccessToken))
	assert.False(t, ok)
	assert.Nil(t, result)
}
