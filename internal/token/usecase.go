package token

import (
	"context"

	"source.toby3d.me/toby3d/auth/internal/domain"
)

type (
	ExchangeOptions struct {
		ClientID     *domain.ClientID
		RedirectURI  *domain.URL
		Code         string
		CodeVerifier string
	}

	UseCase interface {
		Exchange(ctx context.Context, opts ExchangeOptions) (*domain.Token, *domain.Profile, error)

		// Verify checks the AccessToken and returns the associated information.
		Verify(ctx context.Context, accessToken string) (*domain.Token, *domain.Profile, error)

		// Revoke revokes the AccessToken and blocks its further use.
		Revoke(ctx context.Context, accessToken string) error
	}
)

var (
	ErrRevoke error = domain.NewError(
		domain.ErrorCodeAccessDenied,
		"this token has been revoked",
		"",
	)
	ErrMismatchClientID error = domain.NewError(
		domain.ErrorCodeInvalidRequest,
		"client's URL MUST match the client_id used in the authentication request",
		"https://indieauth.net/source/#request",
	)
	ErrMismatchRedirectURI error = domain.NewError(
		domain.ErrorCodeInvalidRequest,
		"client's redirect URL MUST match the initial authentication request",
		"https://indieauth.net/source/#request",
	)
	ErrEmptyScope error = domain.NewError(
		domain.ErrorCodeInvalidScope,
		"empty scopes are invalid",
		"",
	)
	ErrMismatchPKCE error = domain.NewError(
		domain.ErrorCodeInvalidRequest,
		"code_verifier is not hashes to the same value as given in the code_challenge in the original "+
			"authorization request",
		"https://indieauth.net/source/#request",
	)
)
