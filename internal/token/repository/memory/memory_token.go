package memory

import (
	"context"
	"sync"

	"source.toby3d.me/website/oauth/internal/model"
	"source.toby3d.me/website/oauth/internal/token"
)

type memoryTokenRepository struct {
	mutex  *sync.RWMutex
	tokens []*model.Token
}

func NewMemoryTokenRepository() token.Repository {
	return &memoryTokenRepository{
		mutex:  new(sync.RWMutex),
		tokens: make([]*model.Token, 0),
	}
}

func (repo *memoryTokenRepository) Create(ctx context.Context, token *model.Token) error {
	repo.mutex.Lock()

	repo.tokens = append(repo.tokens, token)

	repo.mutex.Unlock()

	return nil
}

func (repo *memoryTokenRepository) Get(ctx context.Context, token string) (*model.Token, error) {
	repo.mutex.RLock()
	defer repo.mutex.RUnlock()

	for i := range repo.tokens {
		if repo.tokens[i].AccessToken != token {
			continue
		}

		return repo.tokens[i], nil
	}

	return nil, nil
}

func (repo *memoryTokenRepository) Delete(ctx context.Context, token string) error {
	repo.mutex.RLock()

	for i := range repo.tokens {
		if repo.tokens[i].AccessToken != token {
			continue
		}

		repo.mutex.RUnlock()
		repo.mutex.Lock()

		if i < len(repo.tokens)-1 {
			copy(repo.tokens[i:], repo.tokens[i+1:])
		}

		repo.tokens[len(repo.tokens)-1] = nil
		repo.tokens = repo.tokens[:len(repo.tokens)-1]

		repo.mutex.Unlock()
		repo.mutex.RLock()

		break
	}

	repo.mutex.RUnlock()

	return nil
}
