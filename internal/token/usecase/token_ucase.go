package usecase

import (
	"context"
	"strings"

	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwt"
	"github.com/pkg/errors"
	"golang.org/x/xerrors"

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
