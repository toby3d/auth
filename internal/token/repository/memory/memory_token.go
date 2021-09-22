package memory

import (
	"context"
	"sync"

	"source.toby3d.me/website/oauth/internal/model"
	"source.toby3d.me/website/oauth/internal/token"
)

type memoryTokenRepository struct {
	tokens *sync.Map
}

func NewMemoryTokenRepository(tokens *sync.Map) token.Repository {
	return &memoryTokenRepository{
		tokens: tokens,
	}
}

func (repo *memoryTokenRepository) Get(ctx context.Context, accessToken string) (*model.Token, error) {
	src, ok := repo.tokens.Load(accessToken)
	if !ok {
		return nil, nil
	}

	result, ok := src.(*model.Token)
	if !ok {
		return nil, nil
	}

	return result, nil
}

func (repo *memoryTokenRepository) Create(ctx context.Context, accessToken *model.Token) error {
	t, err := repo.Get(ctx, accessToken.AccessToken)
	if err != nil {
		return err
	}

	if t != nil {
		return token.ErrExist
	}

	return repo.Update(ctx, accessToken)
}

func (repo *memoryTokenRepository) Update(ctx context.Context, accessToken *model.Token) error {
	repo.tokens.Store(accessToken.AccessToken, accessToken)

	return nil
}

func (repo *memoryTokenRepository) Remove(ctx context.Context, accessToken string) error {
	repo.tokens.Delete(accessToken)

	return nil
}
