package auth

import (
	"context"

	"source.toby3d.me/website/indieauth/internal/domain"
)

type (
	GenerateOptions struct {
		ClientID            *domain.ClientID
		Me                  *domain.Me
		RedirectURI         *domain.URL
		CodeChallengeMethod domain.CodeChallengeMethod
		Scope               domain.Scopes
		CodeChallenge       string
	}

	ExchangeOptions struct {
		ClientID     *domain.ClientID
		RedirectURI  *domain.URL
		Code         string
		CodeVerifier string
	}

	UseCase interface {
		Generate(ctx context.Context, opts GenerateOptions) (string, error)
		Exchange(ctx context.Context, opts ExchangeOptions) (*domain.Me, error)
	}
)
