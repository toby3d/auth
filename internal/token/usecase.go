package token

import (
	"context"
	"errors"

	"source.toby3d.me/website/indieauth/internal/domain"
)

type (
	ExchangeOptions struct {
		ClientID     *domain.ClientID
		RedirectURI  *domain.URL
		Code         string
		CodeVerifier string
	}

	UseCase interface {
		Exchange(ctx context.Context, opts ExchangeOptions) (*domain.Token, error)

		// Verify checks the AccessToken and returns the associated information.
		Verify(ctx context.Context, accessToken string) (*domain.Token, error)

		// Revoke revokes the AccessToken and blocks its further use.
		Revoke(ctx context.Context, accessToken string) error
	}
)

var ErrRevoke = errors.New("this token has been revoked")
