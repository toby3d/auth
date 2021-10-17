package usecase_test

import (
	"context"
	"sync"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	configrepo "source.toby3d.me/website/oauth/internal/config/repository/viper"
	configucase "source.toby3d.me/website/oauth/internal/config/usecase"
	"source.toby3d.me/website/oauth/internal/domain"
	"source.toby3d.me/website/oauth/internal/token"
	repository "source.toby3d.me/website/oauth/internal/token/repository/memory"
	ucase "source.toby3d.me/website/oauth/internal/token/usecase"
)

func TestVerify(t *testing.T) {
	t.Parallel()

	v := viper.New()
	v.Set("indieauth.jwtSigningAlgorithm", "HS256")
	v.Set("indieauth.jwtSecret", "hackme")

	repo := repository.NewMemoryTokenRepository(new(sync.Map))
	useCase := ucase.NewTokenUseCase(repo, configucase.NewConfigUseCase(configrepo.NewViperConfigRepository(v)))

	t.Run("valid", func(t *testing.T) {
		t.Parallel()

		accessToken := domain.TestToken(t)

		result, err := useCase.Verify(context.TODO(), accessToken.AccessToken)
		require.NoError(t, err)
		assert.Equal(t, accessToken, result)
	})

	t.Run("revoke", func(t *testing.T) {
		t.Parallel()

		accessToken := domain.TestToken(t)
		require.NoError(t, repo.Create(context.TODO(), accessToken))

		result, err := useCase.Verify(context.TODO(), accessToken.AccessToken)
		require.ErrorIs(t, err, token.ErrRevoke)
		assert.Nil(t, result)
	})
}

func TestRevoke(t *testing.T) {
	t.Parallel()

	v := viper.New()
	v.Set("indieauth.jwtSigningAlgorithm", "HS256")
	v.Set("indieauth.jwtSecret", "hackme")

	repo := repository.NewMemoryTokenRepository(new(sync.Map))
	accessToken := domain.TestToken(t)

	require.NoError(t, ucase.NewTokenUseCase(
		repo, configucase.NewConfigUseCase(configrepo.NewViperConfigRepository(v)),
	).Revoke(context.TODO(), accessToken.AccessToken))

	result, err := repo.Get(context.TODO(), accessToken.AccessToken)
	assert.NoError(t, err)
	assert.Equal(t, accessToken.AccessToken, result.AccessToken)
}
