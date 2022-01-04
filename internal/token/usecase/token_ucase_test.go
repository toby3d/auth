package usecase_test

import (
	"context"
	"strings"
	"sync"
	"testing"

	"github.com/lestrrat-go/jwx/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	configrepo "source.toby3d.me/website/indieauth/internal/config/repository/viper"
	configucase "source.toby3d.me/website/indieauth/internal/config/usecase"
	"source.toby3d.me/website/indieauth/internal/domain"
	"source.toby3d.me/website/indieauth/internal/token"
	repository "source.toby3d.me/website/indieauth/internal/token/repository/memory"
	ucase "source.toby3d.me/website/indieauth/internal/token/usecase"
)

func TestGenerate(t *testing.T) {
	t.Parallel()

	configer := configucase.NewConfigUseCase(configrepo.NewViperConfigRepository(domain.TestConfig(t)))
	options := token.GenerateOptions{
		ClientID:    "https://app.example.com/",
		Me:          "https://user.example.net/",
		Scopes:      []string{"create", "update", "delete"},
		NonceLength: 42,
	}

	result, err := ucase.NewTokenUseCase(ucase.Config{
		Configer: configer,
		Tokens:   nil,
	}).Generate(context.TODO(), options)
	require.NoError(t, err)
	assert.Equal(t, options.ClientID, result.ClientID)
	assert.Equal(t, options.Me, result.Me)
	assert.Equal(t, options.Scopes, result.Scopes)

	token, err := jwt.ParseString(result.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, options.Me, token.Subject())
	assert.Equal(t, options.ClientID, token.Issuer())

	scope, ok := token.Get("scope")
	require.True(t, ok)
	assert.Equal(t, strings.Join(options.Scopes, " "), scope)
}

func TestVerify(t *testing.T) {
	t.Parallel()

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
