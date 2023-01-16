package memory

import (
	"context"
	"sync"

	"source.toby3d.me/toby3d/auth/internal/domain"
	"source.toby3d.me/toby3d/auth/internal/user"
)

type memoryUserRepository struct {
	mutex *sync.RWMutex
	users map[string]domain.User
}

func NewMemoryUserRepository() user.Repository {
	return &memoryUserRepository{
		mutex: new(sync.RWMutex),
		users: make(map[string]domain.User),
	}
}

func (repo *memoryUserRepository) Create(ctx context.Context, user domain.User) error {
	repo.mutex.Lock()
	defer repo.mutex.Unlock()

	repo.users[user.Me.String()] = user

	return nil
}

func (repo *memoryUserRepository) Get(ctx context.Context, me domain.Me) (*domain.User, error) {
	repo.mutex.RLock()
	defer repo.mutex.RUnlock()

	if u, ok := repo.users[me.String()]; ok {
		return &u, nil
	}

	return nil, user.ErrNotExist
}
