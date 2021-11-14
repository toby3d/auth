package token

import (
	"context"
	"errors"

	"source.toby3d.me/website/oauth/internal/domain"
)

type (
	GenerateOptions struct {
		ClientID    string
		Me          string
		Scopes      []string
		NonceLength int
	}

	UseCase interface {
		// Generate generates a new Token based on the session data.
		Generate(ctx context.Context, opts GenerateOptions) (*domain.Token, error)

		// Verify checks the AccessToken and returns the associated information.
		Verify(ctx context.Context, accessToken string) (*domain.Token, error)

		// Revoke revokes the AccessToken and blocks its further use.
		Revoke(ctx context.Context, accessToken string) error
	}
)

var ErrRevoke = errors.New("this token has been revoked")
