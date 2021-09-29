package usecase

import (
	"context"

	"github.com/pkg/errors"

	"source.toby3d.me/website/oauth/internal/domain"
	"source.toby3d.me/website/oauth/internal/token"
)

type tokenUseCase struct {
	tokens token.Repository
}

func NewTokenUseCase(tokens token.Repository) token.UseCase {
	return &tokenUseCase{
		tokens: tokens,
	}
}

func (useCase *tokenUseCase) Verify(ctx context.Context, accessToken string) (*domain.Token, error) {
	t, err := useCase.tokens.Get(ctx, accessToken)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve token from storage")
	}

	return t, nil
}

func (useCase *tokenUseCase) Revoke(ctx context.Context, accessToken string) error {
	if err := useCase.tokens.Remove(ctx, accessToken); err != nil {
		return errors.Wrap(err, "failed to delete a token in the vault")
	}

	return nil
}
