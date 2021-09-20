package usecase_test

import (
	"context"
	"testing"

	"github.com/brianvoe/gofakeit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"source.toby3d.me/website/oauth/internal/model"
	repository "source.toby3d.me/website/oauth/internal/token/repository/memory"
	"source.toby3d.me/website/oauth/internal/token/usecase"
)

func TestRevoke(t *testing.T) {
	t.Parallel()

	require := require.New(t)
	assert := assert.New(t)
	repo := repository.NewMemoryTokenRepository()
	ucase := usecase.NewTokenUseCase(repo)
	accessToken := gofakeit.Password(true, true, true, true, false, 32)

	require.NoError(repo.Create(context.TODO(), &model.Token{
		AccessToken: accessToken,
		Type:        "Bearer",
		ClientID:    "https://app.example.com/",
		Me:          "https://user.example.net/",
		Scopes:      []string{"create", "update", "delete"},
	}))

	token, err := repo.Get(context.TODO(), accessToken)
	require.NoError(err)
	assert.NotNil(token)

	require.NoError(ucase.Revoke(context.TODO(), token.AccessToken))

	token, err = repo.Get(context.TODO(), token.AccessToken)
	require.NoError(err)
	assert.Nil(token)
}
