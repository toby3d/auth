package usecase

import (
	"context"
	"fmt"

	"golang.org/x/xerrors"

	"source.toby3d.me/website/indieauth/internal/auth"
	"source.toby3d.me/website/indieauth/internal/domain"
	"source.toby3d.me/website/indieauth/internal/random"
	"source.toby3d.me/website/indieauth/internal/session"
)

type authUseCase struct {
	config   *domain.Config
	sessions session.Repository
}

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
		return nil, err
	}

	if opts.ClientID.String() != session.ClientID.String() {
		return nil, domain.Error{
			Code:        "invalid_request",
			Description: "client's URL MUST match the client_id used in the authentication request",
			URI:         "https://indieauth.net/source/#request",
			Frame:       xerrors.Caller(1),
		}
	}

	if opts.RedirectURI.String() != session.RedirectURI.String() {
		return nil, domain.Error{
			Code:        "invalid_request",
			Description: "client's redirect URL MUST match the initial authentication request",
			URI:         "https://indieauth.net/source/#request",
			Frame:       xerrors.Caller(1),
		}
	}

	if session.CodeChallenge != "" &&
		!session.CodeChallengeMethod.Validate(session.CodeChallenge, opts.CodeVerifier) {
		return nil, domain.Error{
			Code: "invalid_request",
			Description: "code_verifier is not hashes to the same value as given in " +
				"the code_challenge in the original authorization request",
			URI:   "https://indieauth.net/source/#request",
			Frame: xerrors.Caller(1),
		}
	}

	return session.Me, nil
}
