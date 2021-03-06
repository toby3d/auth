package memory

import (
	"context"
	"errors"
	"path"
	"sync"

	"source.toby3d.me/toby3d/auth/internal/domain"
	"source.toby3d.me/toby3d/auth/internal/token"
)

type memoryTokenRepository struct {
	store *sync.Map
}

const DefaultPathPrefix string = "tokens"

func NewMemoryTokenRepository(store *sync.Map) token.Repository {
	return &memoryTokenRepository{
		store: store,
	}
}

func (repo *memoryTokenRepository) Create(ctx context.Context, accessToken *domain.Token) error {
	t, err := repo.Get(ctx, accessToken.AccessToken)
	if err != nil && !errors.Is(err, token.ErrNotExist) {
		return err
	}

	if t != nil {
		return token.ErrExist
	}

	repo.store.Store(path.Join(DefaultPathPrefix, accessToken.AccessToken), accessToken)

	return nil
}

func (repo *memoryTokenRepository) Get(ctx context.Context, accessToken string) (*domain.Token, error) {
	t, ok := repo.store.Load(path.Join(DefaultPathPrefix, accessToken))
	if !ok {
		return nil, token.ErrNotExist
	}

	result, ok := t.(*domain.Token)
	if !ok {
		return nil, token.ErrNotExist
	}

	return result, nil
}
