package auth

import (
	"context"
	"net/url"

	"source.toby3d.me/toby3d/auth/internal/domain"
)

type (
	GenerateOptions struct {
		ClientID            *domain.ClientID
		Me                  *domain.Me
		RedirectURI         *url.URL
		CodeChallengeMethod domain.CodeChallengeMethod
		Scope               domain.Scopes
		CodeChallenge       string
	}

	ExchangeOptions struct {
		ClientID     *domain.ClientID
		RedirectURI  *url.URL
		Code         string
		CodeVerifier string
	}

	UseCase interface {
		Generate(ctx context.Context, opts GenerateOptions) (string, error)
		Exchange(ctx context.Context, opts ExchangeOptions) (*domain.Me, *domain.Profile, error)
	}
)

var (
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
	ErrMismatchPKCE error = domain.NewError(
		domain.ErrorCodeInvalidRequest,
		"code_verifier is not hashes to the same value as given in the code_challenge in the original "+
			" authorization request",
		"https://indieauth.net/source/#request",
	)
)
