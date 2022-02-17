package usecase

import (
	"context"
	"fmt"

	"source.toby3d.me/website/indieauth/internal/auth"
	"source.toby3d.me/website/indieauth/internal/domain"
	"source.toby3d.me/website/indieauth/internal/profile"
	"source.toby3d.me/website/indieauth/internal/random"
	"source.toby3d.me/website/indieauth/internal/session"
)

type authUseCase struct {
	config   *domain.Config
	sessions session.Repository
	profiles profile.Repository
}

// NewAuthUseCase creates a new authentication use case.
func NewAuthUseCase(sessions session.Repository, profiles profile.Repository, config *domain.Config) auth.UseCase {
	return &authUseCase{
		config:   config,
		sessions: sessions,
		profiles: profiles,
	}
}

func (uc *authUseCase) Generate(ctx context.Context, opts auth.GenerateOptions) (string, error) {
	code, err := random.String(uc.config.Code.Length)
	if err != nil {
		return "", fmt.Errorf("cannot generate random code: %w", err)
	}

	var userInfo *domain.Profile

	// NOTE(toby3d): We request information about the profile only if there
	// is a corresponding Scope. However, the availability of this
	// information in the token is not guaranteed and is completely optional:
	// https://indieauth.net/source/#profile-information
	if opts.Scope.Has(domain.ScopeProfile) {
		if userInfo, err = uc.profiles.Get(ctx, opts.Me); err == nil &&
			userInfo.Email != nil && !opts.Scope.Has(domain.ScopeEmail) {
			userInfo.Email = nil
		}
	}

	if err = uc.sessions.Create(ctx, &domain.Session{
		ClientID:            opts.ClientID,
		Code:                code,
		CodeChallenge:       opts.CodeChallenge,
		CodeChallengeMethod: opts.CodeChallengeMethod,
		Me:                  opts.Me,
		Profile:             userInfo,
		RedirectURI:         opts.RedirectURI,
		Scope:               opts.Scope,
	}); err != nil {
		return "", fmt.Errorf("cannot save session in store: %w", err)
	}

	return code, nil
}

func (uc *authUseCase) Exchange(ctx context.Context, opts auth.ExchangeOptions) (*domain.Me, *domain.Profile, error) {
	session, err := uc.sessions.GetAndDelete(ctx, opts.Code)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot find session in store: %w", err)
	}

	if opts.ClientID.String() != session.ClientID.String() {
		return nil, nil, auth.ErrMismatchClientID
	}

	if opts.RedirectURI.String() != session.RedirectURI.String() {
		return nil, nil, auth.ErrMismatchRedirectURI
	}

	if session.CodeChallenge != "" &&
		session.CodeChallengeMethod != domain.CodeChallengeMethodUndefined &&
		!session.CodeChallengeMethod.Validate(session.CodeChallenge, opts.CodeVerifier) {
		return nil, nil, auth.ErrMismatchPKCE
	}

	return session.Me, session.Profile, nil
}
