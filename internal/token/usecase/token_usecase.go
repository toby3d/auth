package usecase

import (
	"context"
	"strings"

	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwt"
	"github.com/pkg/errors"

	"source.toby3d.me/website/oauth/internal/config"
	"source.toby3d.me/website/oauth/internal/domain"
	"source.toby3d.me/website/oauth/internal/token"
)

type tokenUseCase struct {
	tokens   token.Repository
	configer config.UseCase
}

func NewTokenUseCase(tokens token.Repository, configer config.UseCase) token.UseCase {
	return &tokenUseCase{
		tokens:   tokens,
		configer: configer,
	}
}

func (useCase *tokenUseCase) Verify(ctx context.Context, accessToken string) (*domain.Token, error) {
	token, err := useCase.tokens.Get(ctx, accessToken)
	if err != nil {
		return nil, errors.Wrap(err, "cannot find token in database")
	}

	if token != nil {
		return nil, nil
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

	token = &domain.Token{
		Expiry:      t.Expiration(),
		Scopes:      []string{},
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ClientID:    t.Issuer(),
		Me:          t.Subject(),
	}

	if scope, ok := t.Get("scope"); ok {
		token.Scopes = strings.Fields(scope.(string))
	}

	return token, nil
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
