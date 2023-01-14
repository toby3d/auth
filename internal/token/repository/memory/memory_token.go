package memory

import (
	"context"
	"sync"

	"source.toby3d.me/toby3d/auth/internal/domain"
	"source.toby3d.me/toby3d/auth/internal/token"
)

type memoryTokenRepository struct {
	mutex  *sync.RWMutex
	tokens map[string]domain.Token
}

func NewMemoryTokenRepository() token.Repository {
	return &memoryTokenRepository{
		mutex:  new(sync.RWMutex),
		tokens: make(map[string]domain.Token),
	}
}

func (repo *memoryTokenRepository) Create(ctx context.Context, accessToken domain.Token) error {
	repo.mutex.Lock()
	defer repo.mutex.Unlock()

	repo.tokens[accessToken.AccessToken] = accessToken

	return nil
}

func (repo *memoryTokenRepository) Get(ctx context.Context, accessToken string) (*domain.Token, error) {
	repo.mutex.RLock()
	defer repo.mutex.RUnlock()

	if t, ok := repo.tokens[accessToken]; ok {
		return &t, nil
	}

	return nil, token.ErrNotExist
}
