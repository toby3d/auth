package memory

import (
	"context"
	"sync"

	"gitlab.com/toby3d/indieauth/internal/model"
	"gitlab.com/toby3d/indieauth/internal/token"
)

type memoryTokenRepository struct {
	tokens *sync.Map
}

func NewMemoryTokenRepository() token.Repository {
	return &memoryTokenRepository{
		tokens: new(sync.Map),
	}
}

func (repo *memoryTokenRepository) Create(ctx context.Context, token *model.Token) error {
	repo.tokens.Store(token.AccessToken, token)

	return nil
}

func (repo *memoryTokenRepository) Delete(ctx context.Context, token string) error {
	repo.tokens.Delete(token)

	return nil
}

func (repo *memoryTokenRepository) Get(ctx context.Context, token string) (*model.Token, error) {
	t, ok := repo.tokens.Load(token)
	if !ok {
		return nil, nil
	}

	return t.(*model.Token), nil
}
