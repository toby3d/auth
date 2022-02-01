package usecase

import (
	"context"
	"fmt"

	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwt"
	"golang.org/x/xerrors"

	"source.toby3d.me/website/indieauth/internal/domain"
	"source.toby3d.me/website/indieauth/internal/session"
	"source.toby3d.me/website/indieauth/internal/token"
)

type tokenUseCase struct {
	sessions session.Repository
	config   *domain.Config
	tokens   token.Repository
}

func NewTokenUseCase(tokens token.Repository, sessions session.Repository, config *domain.Config) token.UseCase {
	jwt.RegisterCustomField("scope", make(domain.Scopes, 0))

	return &tokenUseCase{
		sessions: sessions,
		config:   config,
		tokens:   tokens,
	}
}

func (useCase *tokenUseCase) Exchange(ctx context.Context, opts token.ExchangeOptions) (*domain.Token, *domain.Profile,
	error) {
	session, err := useCase.sessions.GetAndDelete(ctx, opts.Code)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot get session from store: %w", err)
	}

	if opts.ClientID.String() != session.ClientID.String() {
		return nil, nil, token.ErrMismatchClientID
	}

	if opts.RedirectURI.String() != session.RedirectURI.String() {
		return nil, nil, token.ErrMismatchRedirectURI
	}

	if session.CodeChallenge != "" &&
		!session.CodeChallengeMethod.Validate(session.CodeChallenge, opts.CodeVerifier) {
		return nil, nil, token.ErrMismatchPKCE
	}

	// NOTE(toby3d): If the authorization code was issued with no scope, the
	// token endpoint MUST NOT issue an access token, as empty scopes are
	// invalid per Section 3.3 of OAuth 2.0 RFC6749.
	if session.Scope.IsEmpty() {
		return nil, nil, token.ErrEmptyScope
	}

	tkn, err := domain.NewToken(domain.NewTokenOptions{
		Algorithm:   useCase.config.JWT.Algorithm,
		Expiration:  useCase.config.JWT.Expiry,
		Issuer:      session.ClientID,
		NonceLength: useCase.config.JWT.NonceLength,
		Scope:       session.Scope,
		Secret:      []byte(useCase.config.JWT.Secret),
		Subject:     session.Me,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("cannot generate a new access token: %w", err)
	}

	if !session.Scope.Has(domain.ScopeProfile) {
		return tkn, nil, nil
	}

	p := new(domain.Profile)

	// TODO(toby3d): if session.Scope.Has(domain.ScopeEmail) {}

	return tkn, p, nil
}

func (useCase *tokenUseCase) Verify(ctx context.Context, accessToken string) (*domain.Token, error) {
	find, err := useCase.tokens.Get(ctx, accessToken)
	if err != nil && !xerrors.Is(err, token.ErrNotExist) {
		return nil, fmt.Errorf("cannot check token in store: %w", err)
	}

	if find != nil {
		return nil, token.ErrRevoke
	}

	tkn, err := jwt.ParseString(accessToken, jwt.WithVerify(jwa.SignatureAlgorithm(useCase.config.JWT.Algorithm),
		[]byte(useCase.config.JWT.Secret)))
	if err != nil {
		return nil, fmt.Errorf("cannot parse JWT token: %w", err)
	}

	if err = jwt.Validate(tkn); err != nil {
		return nil, fmt.Errorf("cannot validate JWT token: %w", err)
	}

	result := &domain.Token{
		Scope:       nil,
		ClientID:    nil,
		Me:          nil,
		AccessToken: accessToken,
	}
	result.ClientID, _ = domain.ParseClientID(tkn.Issuer())
	result.Me, _ = domain.ParseMe(tkn.Subject())

	if scope, ok := tkn.Get("scope"); ok {
		result.Scope, _ = scope.(domain.Scopes)
	}

	return result, nil
}

func (useCase *tokenUseCase) Revoke(ctx context.Context, accessToken string) error {
	tkn, err := useCase.Verify(ctx, accessToken)
	if err != nil {
		return fmt.Errorf("cannot verify token: %w", err)
	}

	if err = useCase.tokens.Create(ctx, tkn); err != nil {
		return fmt.Errorf("cannot save token in database: %w", err)
	}

	return nil
}
