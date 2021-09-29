package usecase_test

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"source.toby3d.me/website/oauth/internal/domain"
	repository "source.toby3d.me/website/oauth/internal/token/repository/memory"
	"source.toby3d.me/website/oauth/internal/token/usecase"
)

func TestVerify(t *testing.T) {
	t.Parallel()

	repo := repository.NewMemoryTokenRepository(new(sync.Map))
	accessToken := domain.NewToken()

	require.NoError(t, repo.Create(context.TODO(), accessToken))

	token, err := usecase.NewTokenUseCase(repo).Verify(context.TODO(), accessToken.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, accessToken, token)
}

func TestRevoke(t *testing.T) {
	t.Parallel()

	repo := repository.NewMemoryTokenRepository(new(sync.Map))
	accessToken := domain.TestToken(t)

	require.NoError(t, repo.Create(context.TODO(), accessToken))

	token, err := repo.Get(context.TODO(), accessToken.AccessToken)
	require.NoError(t, err)
	assert.NotNil(t, token)

	require.NoError(t, usecase.NewTokenUseCase(repo).Revoke(context.TODO(), token.AccessToken))

	token, err = repo.Get(context.TODO(), token.AccessToken)
	require.NoError(t, err)
	assert.Nil(t, token)
}
