package usecase

import (
	"context"
	"fmt"

	"source.toby3d.me/website/indieauth/internal/auth"
	"source.toby3d.me/website/indieauth/internal/domain"
	"source.toby3d.me/website/indieauth/internal/random"
	"source.toby3d.me/website/indieauth/internal/session"
)

type authUseCase struct {
	config   *domain.Config
	sessions session.Repository
}

// NewAuthUseCase creates a new authentication use case.
func NewAuthUseCase(sessions session.Repository, config *domain.Config) auth.UseCase {
	return &authUseCase{
		config:   config,
		sessions: sessions,
	}
}

func (useCase *authUseCase) Generate(ctx context.Context, opts auth.GenerateOptions) (string, error) {
	code, err := random.String(useCase.config.Code.Length)
	if err != nil {
		return "", fmt.Errorf("cannot generate random code: %w", err)
	}

	if err = useCase.sessions.Create(ctx, &domain.Session{
		ClientID:            opts.ClientID,
		Code:                code,
		CodeChallenge:       opts.CodeChallenge,
		CodeChallengeMethod: opts.CodeChallengeMethod,
		Me:                  opts.Me,
		RedirectURI:         opts.RedirectURI,
		Scope:               opts.Scope,
	}); err != nil {
		return "", fmt.Errorf("cannot save session in store: %w", err)
	}

	return code, nil
}

func (useCase *authUseCase) Exchange(ctx context.Context, opts auth.ExchangeOptions) (*domain.Me, error) {
	session, err := useCase.sessions.GetAndDelete(ctx, opts.Code)
	if err != nil {
		return nil, fmt.Errorf("cannot find session in store: %w", err)
	}

	if opts.ClientID.String() != session.ClientID.String() {
		return nil, auth.ErrMismatchClientID
	}

	if opts.RedirectURI.String() != session.RedirectURI.String() {
		return nil, auth.ErrMismatchRedirectURI
	}

	if session.CodeChallenge != "" && session.CodeChallengeMethod != domain.CodeChallengeMethodUndefined &&
		!session.CodeChallengeMethod.Validate(session.CodeChallenge, opts.CodeVerifier) {
		return nil, auth.ErrMismatchPKCE
	}

	return session.Me, nil
}
