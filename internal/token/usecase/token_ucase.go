package usecase

import (
	"context"
	"strings"
	"time"

	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwt"
	"github.com/pkg/errors"
	"golang.org/x/xerrors"

	"source.toby3d.me/website/oauth/internal/config"
	"source.toby3d.me/website/oauth/internal/domain"
	"source.toby3d.me/website/oauth/internal/random"
	"source.toby3d.me/website/oauth/internal/token"
)

type (
	Config struct {
		Configer config.UseCase
		Tokens   token.Repository
	}

	tokenUseCase struct {
		configer config.UseCase
		tokens   token.Repository
	}
)

func NewTokenUseCase(config Config) token.UseCase {
	return &tokenUseCase{
		configer: config.Configer,
		tokens:   config.Tokens,
	}
}

// Generate generates a new Token based on the session data.
func (useCase *tokenUseCase) Generate(ctx context.Context, opts token.GenerateOptions) (*domain.Token, error) {
	nonce, err := random.String(opts.NonceLength)
	if err != nil {
		return nil, errors.Wrap(err, "cannot generate code")
	}

	t := jwt.New()
	now := time.Now().UTC().Round(time.Second)

	t.Set(jwt.IssuerKey, opts.ClientID)
	t.Set(jwt.SubjectKey, opts.Me)
	t.Set(jwt.ExpirationKey, now.Add(useCase.configer.GetIndieAuthAccessTokenExpirationTime()))
	t.Set(jwt.NotBeforeKey, now)
	t.Set(jwt.IssuedAtKey, now)
	t.Set("scope", strings.Join(opts.Scopes, " "))
	t.Set("nonce", nonce)

	token, err := jwt.Sign(t,
		jwa.SignatureAlgorithm(useCase.configer.GetIndieAuthJWTSigningAlgorithm()),
		[]byte(useCase.configer.GetIndieAuthJWTSecret()))
	if err != nil {
		return nil, errors.Wrap(err, "cannot sign a new access token")
	}

	return &domain.Token{
		Scopes:      opts.Scopes,
		AccessToken: string(token),
		ClientID:    opts.ClientID,
		Me:          opts.Me,
	}, nil
}

func (useCase *tokenUseCase) Verify(ctx context.Context, accessToken string) (*domain.Token, error) {
	find, err := useCase.tokens.Get(ctx, accessToken)
	if err != nil && !xerrors.Is(err, token.ErrNotExist) {
		return nil, errors.Wrap(err, "cannot ckeck token in store")
	}

	if find != nil {
		return nil, token.ErrRevoke
	}

	t, err := jwt.ParseString(accessToken, jwt.WithVerify(
		jwa.SignatureAlgorithm(useCase.configer.GetIndieAuthJWTSigningAlgorithm()),
		[]byte(useCase.configer.GetIndieAuthJWTSecret()),
	))
	if err != nil {
		return nil, errors.Wrap(err, "cannot parse JWT token")
	}

	if err = jwt.Validate(t); err != nil {
		return nil, errors.Wrap(err, "cannot validate JWT token")
	}

	result := &domain.Token{
		AccessToken: accessToken,
		ClientID:    t.Issuer(),
		Me:          t.Subject(),
		Scopes:      make([]string, 0),
	}

	rawScope, ok := t.Get("scope")
	if !ok {
		return result, nil
	}

	if scope, ok := rawScope.(string); ok {
		result.Scopes = strings.Fields(scope)
	}

	return result, nil
}

func (useCase *tokenUseCase) Revoke(ctx context.Context, accessToken string) error {
	t, err := useCase.Verify(ctx, accessToken)
	if err != nil {
		return errors.Wrap(err, "cannot verify token")
	}

	if err = useCase.tokens.Create(ctx, t); err != nil {
		return errors.Wrap(err, "cannot save token in database")
	}

	return nil
}
