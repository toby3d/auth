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

//nolint: gochecknoinits
func init() {
	jwt.RegisterCustomField("scope", make(domain.Scopes, 0))
}

func NewTokenUseCase(tokens token.Repository, sessions session.Repository, config *domain.Config) token.UseCase {
	return &tokenUseCase{
		sessions: sessions,
		config:   config,
		tokens:   tokens,
	}
}

func (useCase *tokenUseCase) Exchange(ctx context.Context, opts token.ExchangeOptions) (*domain.Token, error) {
	session, err := useCase.sessions.GetAndDelete(ctx, opts.Code)
	if err != nil {
		return nil, fmt.Errorf("cannot get session from store: %w", err)
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

	t, err := domain.NewToken(domain.NewTokenOptions{
		Algorithm:   useCase.config.JWT.Algorithm,
		Expiration:  useCase.config.JWT.Expiry,
		Issuer:      session.ClientID,
		NonceLength: useCase.config.JWT.NonceLength,
		Scope:       session.Scope,
		Secret:      []byte(useCase.config.JWT.Secret),
		Subject:     session.Me,
	})
	if err != nil {
		return nil, fmt.Errorf("cannot generate a new access token: %w", err)
	}

	return t, nil
}

func (useCase *tokenUseCase) Verify(ctx context.Context, accessToken string) (*domain.Token, error) {
	find, err := useCase.tokens.Get(ctx, accessToken)
	if err != nil && !xerrors.Is(err, token.ErrNotExist) {
		return nil, fmt.Errorf("cannot check token in store: %w", err)
	}

	if find != nil {
		return nil, token.ErrRevoke
	}

	t, err := jwt.ParseString(accessToken, jwt.WithVerify(jwa.SignatureAlgorithm(useCase.config.JWT.Algorithm),
		[]byte(useCase.config.JWT.Secret)))
	if err != nil {
		return nil, fmt.Errorf("cannot parse JWT token: %w", err)
	}

	if err = jwt.Validate(t); err != nil {
		return nil, fmt.Errorf("cannot validate JWT token: %w", err)
	}

	result := &domain.Token{
		AccessToken: accessToken,
	}
	result.ClientID, _ = domain.ParseClientID(t.Issuer())
	result.Me, _ = domain.NewMe(t.Subject())

	if scope, ok := t.Get("scope"); ok {
		result.Scope, _ = scope.(domain.Scopes)
	}

	return result, nil
}

func (useCase *tokenUseCase) Revoke(ctx context.Context, accessToken string) error {
	t, err := useCase.Verify(ctx, accessToken)
	if err != nil {
		return fmt.Errorf("cannot verify token: %w", err)
	}

	if err = useCase.tokens.Create(ctx, t); err != nil {
		return fmt.Errorf("cannot save token in database: %w", err)
	}

	return nil
}
