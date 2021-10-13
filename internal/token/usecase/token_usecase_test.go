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
	repository "source.toby3d.me/website/oauth/internal/token/repository/memory"
	"source.toby3d.me/website/oauth/internal/token/usecase"
)

func TestVerify(t *testing.T) {
	t.Parallel()

	v := viper.New()
	v.SetDefault("indieauth.jwtSigningAlgorithm", "HS256")
	v.SetDefault("indieauth.jwtSecret", "hackme")

	repo := repository.NewMemoryTokenRepository(new(sync.Map))
	accessToken := domain.TestToken(t)

	token, err := usecase.NewTokenUseCase(
		repo, configucase.NewConfigUseCase(configrepo.NewViperConfigRepository(v)),
	).Verify(context.TODO(), accessToken.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, accessToken.AccessToken, token.AccessToken)
}

func TestRevoke(t *testing.T) {
	t.Parallel()

	v := viper.New()
	v.SetDefault("indieauth.jwtSigningAlgorithm", "HS256")
	v.SetDefault("indieauth.jwtSecret", "hackme")

	repo := repository.NewMemoryTokenRepository(new(sync.Map))
	accessToken := domain.TestToken(t)

	require.NoError(t, usecase.NewTokenUseCase(
		repo, configucase.NewConfigUseCase(configrepo.NewViperConfigRepository(v)),
	).Revoke(context.TODO(), accessToken.AccessToken))

	result, err := repo.Get(context.TODO(), accessToken.AccessToken)
	assert.NoError(t, err)
	assert.Equal(t, accessToken.AccessToken, result.AccessToken)
}
