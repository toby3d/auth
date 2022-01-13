package auth

import (
	"context"

	"source.toby3d.me/website/indieauth/internal/domain"
)

type (
	GenerateOptions struct {
		ClientID            *domain.ClientID
		RedirectURI         *domain.URL
		CodeChallenge       string
		CodeChallengeMethod domain.CodeChallengeMethod
		Scope               domain.Scopes
		Me                  *domain.Me
	}

	ExchangeOptions struct {
		Code         string
		ClientID     *domain.ClientID
		RedirectURI  *domain.URL
		CodeVerifier string
	}

	UseCase interface {
		Generate(ctx context.Context, opts GenerateOptions) (string, error)
		Exchange(ctx context.Context, opts ExchangeOptions) (*domain.Me, error)
	}
)
