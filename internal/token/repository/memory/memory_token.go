package memory

import (
	"context"
	"path"
	"sync"

	"source.toby3d.me/website/oauth/internal/domain"
	"source.toby3d.me/website/oauth/internal/token"
)

type memoryTokenRepository struct {
	tokens *sync.Map
}

const Key string = "tokens"

func NewMemoryTokenRepository(tokens *sync.Map) token.Repository {
	return &memoryTokenRepository{
		tokens: tokens,
	}
}

func (repo *memoryTokenRepository) Get(ctx context.Context, accessToken string) (*domain.Token, error) {
	src, ok := repo.tokens.Load(path.Join(Key, accessToken))
	if !ok {
		return nil, nil
	}

	result, ok := src.(*domain.Token)
	if !ok {
		return nil, nil
	}

	return result, nil
}

func (repo *memoryTokenRepository) Create(ctx context.Context, accessToken *domain.Token) error {
	t, err := repo.Get(ctx, accessToken.AccessToken)
	if err != nil {
		return err
	}

	if t != nil {
		return token.ErrExist
	}

	return repo.Update(ctx, accessToken)
}

func (repo *memoryTokenRepository) Update(ctx context.Context, accessToken *domain.Token) error {
	repo.tokens.Store(path.Join(Key, accessToken.AccessToken), accessToken)

	return nil
}

func (repo *memoryTokenRepository) Remove(ctx context.Context, accessToken string) error {
	repo.tokens.Delete(path.Join(Key, accessToken))

	return nil
}
