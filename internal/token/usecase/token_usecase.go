package usecase

import (
	"context"

	"source.toby3d.me/website/oauth/internal/model"
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

func (useCase *tokenUseCase) Verify(ctx context.Context, accessToken string) (*model.Token, error) {
	return useCase.tokens.Get(ctx, accessToken)
}

func (useCase *tokenUseCase) Revoke(ctx context.Context, accessToken string) error {
	return useCase.tokens.Remove(ctx, accessToken)
}
