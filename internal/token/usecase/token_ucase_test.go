package usecase_test

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"source.toby3d.me/website/indieauth/internal/domain"
	"source.toby3d.me/website/indieauth/internal/token"
	repository "source.toby3d.me/website/indieauth/internal/token/repository/memory"
	usecase "source.toby3d.me/website/indieauth/internal/token/usecase"
)

func TestExchange(t *testing.T) {
	t.Parallel()
}

func TestVerify(t *testing.T) {
	t.Parallel()

	repo := repository.NewMemoryTokenRepository(new(sync.Map))
	ucase := usecase.NewTokenUseCase(repo, nil, domain.TestConfig(t))

	t.Run("valid", func(t *testing.T) {
		t.Parallel()

		accessToken := domain.TestToken(t)

		result, err := ucase.Verify(context.TODO(), accessToken.AccessToken)
		require.NoError(t, err)
		assert.Equal(t, accessToken.AccessToken, result.AccessToken)
		assert.Equal(t, accessToken.Scope, result.Scope)
		assert.Equal(t, accessToken.ClientID.String(), result.ClientID.String())
		assert.Equal(t, accessToken.Me.String(), result.Me.String())
	})

	t.Run("revoked", func(t *testing.T) {
		t.Parallel()

		accessToken := domain.TestToken(t)
		require.NoError(t, repo.Create(context.TODO(), accessToken))

		result, err := ucase.Verify(context.TODO(), accessToken.AccessToken)
		require.ErrorIs(t, err, token.ErrRevoke)
		assert.Nil(t, result)
	})
}

func TestRevoke(t *testing.T) {
	t.Parallel()

	config := domain.TestConfig(t)
	accessToken := domain.TestToken(t)
	repo := repository.NewMemoryTokenRepository(new(sync.Map))

	require.NoError(t, usecase.NewTokenUseCase(repo, nil, config).
		Revoke(context.TODO(), accessToken.AccessToken))

	result, err := repo.Get(context.TODO(), accessToken.AccessToken)
	assert.NoError(t, err)
	assert.Equal(t, accessToken.AccessToken, result.AccessToken)
}
