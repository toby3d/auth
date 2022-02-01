package usecase_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"source.toby3d.me/website/indieauth/internal/domain"
	"source.toby3d.me/website/indieauth/internal/token"
	repository "source.toby3d.me/website/indieauth/internal/token/repository/memory"
	usecase "source.toby3d.me/website/indieauth/internal/token/usecase"
)

/* TODO(toby3d)
func TestExchange(t *testing.T) {
	t.Parallel()
}
*/

func TestVerify(t *testing.T) {
	t.Parallel()

	repo := repository.NewMemoryTokenRepository(new(sync.Map))
	ucase := usecase.NewTokenUseCase(repo, nil, domain.TestConfig(t))

	t.Run("valid", func(t *testing.T) {
		t.Parallel()

		accessToken := domain.TestToken(t)

		result, err := ucase.Verify(context.TODO(), accessToken.AccessToken)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, accessToken.AccessToken, result.AccessToken)
		assert.Equal(t, accessToken.Scope, result.Scope)
		assert.Equal(t, accessToken.ClientID.String(), result.ClientID.String())
		assert.Equal(t, accessToken.Me.String(), result.Me.String())
	})

	t.Run("revoked", func(t *testing.T) {
		t.Parallel()

		accessToken := domain.TestToken(t)
		if err := repo.Create(context.TODO(), accessToken); err != nil {
			t.Fatal(err)
		}

		result, err := ucase.Verify(context.TODO(), accessToken.AccessToken)
		if !errors.Is(err, token.ErrRevoke) {
			t.Errorf("Verify(%s) = %v, want %v", accessToken.AccessToken, err, token.ErrRevoke)
		}

		if result != nil {
			t.Errorf("Verify(%s) = %v, want %v", accessToken.AccessToken, result, nil)
		}
	})
}

func TestRevoke(t *testing.T) {
	t.Parallel()

	config := domain.TestConfig(t)
	accessToken := domain.TestToken(t)
	repo := repository.NewMemoryTokenRepository(new(sync.Map))

	if err := usecase.NewTokenUseCase(repo, nil, config).
		Revoke(context.TODO(), accessToken.AccessToken); err != nil {
		t.Fatal(err)
	}

	result, err := repo.Get(context.TODO(), accessToken.AccessToken)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, accessToken.AccessToken, result.AccessToken)
}
