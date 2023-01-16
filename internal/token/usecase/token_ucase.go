package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwt"

	"source.toby3d.me/toby3d/auth/internal/domain"
	"source.toby3d.me/toby3d/auth/internal/profile"
	"source.toby3d.me/toby3d/auth/internal/session"
	"source.toby3d.me/toby3d/auth/internal/token"
)

type (
	Config struct {
		Config   *domain.Config
		Profiles profile.Repository
		Sessions session.Repository
		Tokens   token.Repository
	}

	tokenUseCase struct {
		config   *domain.Config
		profiles profile.Repository
		sessions session.Repository
		tokens   token.Repository
	}
)

func NewTokenUseCase(config Config) token.UseCase {
	jwt.RegisterCustomField("scope", make(domain.Scopes, 0))

	return &tokenUseCase{
		config:   config.Config,
		profiles: config.Profiles,
		sessions: config.Sessions,
		tokens:   config.Tokens,
	}
}

//nolint:cyclop
func (uc *tokenUseCase) Exchange(ctx context.Context, opts token.ExchangeOptions) (*domain.Token, *domain.Profile,
	error,
) {
	session, err := uc.sessions.GetAndDelete(ctx, opts.Code)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot get session from store: %w", err)
	}

	if opts.ClientID.String() != session.ClientID.String() {
		return nil, nil, token.ErrMismatchClientID
	}

	if opts.RedirectURI.String() != session.RedirectURI.String() {
		return nil, nil, token.ErrMismatchRedirectURI
	}

	if session.CodeChallenge != "" && session.CodeChallengeMethod != domain.CodeChallengeMethodUnd &&
		!session.CodeChallengeMethod.Validate(session.CodeChallenge, opts.CodeVerifier) {
		return nil, nil, token.ErrMismatchPKCE
	}

	// NOTE(toby3d): If the authorization code was issued with no scope, the
	// token endpoint MUST NOT issue an access token, as empty scopes are
	// invalid (RFC 6749 section 3.3).
	if session.Scope.IsEmpty() {
		return nil, nil, token.ErrEmptyScope
	}

	if !session.Scope.Has(domain.ScopeProfile) {
		session.Profile = nil
	} else if !session.Scope.Has(domain.ScopeEmail) {
		session.Profile.Email = nil
	}

	tkn, err := domain.NewToken(domain.NewTokenOptions{
		Expiration:  uc.config.JWT.Expiry,
		Issuer:      session.ClientID,
		Subject:     session.Me,
		Scope:       session.Scope,
		Secret:      []byte(uc.config.JWT.Secret),
		Algorithm:   uc.config.JWT.Algorithm,
		NonceLength: uc.config.JWT.NonceLength,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("cannot generate a new access token: %w", err)
	}

	return tkn, session.Profile, nil
}

func (uc *tokenUseCase) Verify(ctx context.Context, accessToken string) (*domain.Token, *domain.Profile, error) {
	if _, err := uc.tokens.Get(ctx, accessToken); err == nil || !errors.Is(err, token.ErrNotExist) {
		return nil, nil, fmt.Errorf("cannot check token in store: %w", err)
	}

	tkn, err := jwt.ParseString(accessToken, jwt.WithKey(jwa.SignatureAlgorithm(uc.config.JWT.Algorithm),
		[]byte(uc.config.JWT.Secret)), jwt.WithVerify(true))
	if err != nil {
		return nil, nil, fmt.Errorf("cannot parse JWT token: %w", err)
	}

	if err = jwt.Validate(tkn); err != nil {
		return nil, nil, fmt.Errorf("cannot validate JWT token: %w", err)
	}

	cid, _ := domain.ParseClientID(tkn.Issuer())
	me, _ := domain.ParseMe(tkn.Subject())
	result := &domain.Token{
		CreatedAt:    tkn.IssuedAt(),
		Expiry:       tkn.Expiration(),
		ClientID:     *cid,
		Me:           *me,
		Scope:        nil,
		AccessToken:  accessToken,
		RefreshToken: "", // TODO(toby3d)
	}

	if scope, ok := tkn.Get("scope"); ok {
		result.Scope, _ = scope.(domain.Scopes)
	}

	if !result.Scope.Has(domain.ScopeProfile) {
		return result, nil, nil
	}

	profile, err := uc.profiles.Get(ctx, result.Me)
	if err != nil {
		return result, nil, nil //nolint:nilerr // it's okay to return result without profile
	}

	if !result.Scope.Has(domain.ScopeEmail) && len(profile.Email) > 0 {
		profile.Email = nil
	}

	return result, profile, nil
}

func (uc *tokenUseCase) Revoke(ctx context.Context, accessToken string) error {
	tkn, _, err := uc.Verify(ctx, accessToken)
	if err != nil {
		if errors.Is(err, token.ErrNotExist) {
			return nil
		}

		return fmt.Errorf("cannot verify token: %w", err)
	}

	if err = uc.tokens.Create(ctx, *tkn); err != nil && !errors.Is(err, token.ErrExist) {
		return fmt.Errorf("cannot save token in database: %w", err)
	}

	return nil
}
