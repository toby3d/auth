package memory_test

import (
	"context"
	"path"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"source.toby3d.me/toby3d/auth/internal/domain"
	"source.toby3d.me/toby3d/auth/internal/token"
	repository "source.toby3d.me/toby3d/auth/internal/token/repository/memory"
)

func TestCreate(t *testing.T) {
	t.Parallel()

	store := new(sync.Map)
	accessToken := domain.TestToken(t)

	repo := repository.NewMemoryTokenRepository(store)
	if err := repo.Create(context.Background(), accessToken); err != nil {
		t.Fatal(err)
	}

	result, ok := store.Load(path.Join(repository.DefaultPathPrefix, accessToken.AccessToken))
	assert.True(t, ok)
	assert.Equal(t, accessToken, result)

	assert.ErrorIs(t, repo.Create(context.Background(), accessToken), token.ErrExist)
}

func TestGet(t *testing.T) {
	t.Parallel()

	store := new(sync.Map)
	accessToken := domain.TestToken(t)

	store.Store(path.Join(repository.DefaultPathPrefix, accessToken.AccessToken), accessToken)

	result, err := repository.NewMemoryTokenRepository(store).Get(context.Background(), accessToken.AccessToken)
	assert.NoError(t, err)
	assert.Equal(t, accessToken, result)
}
