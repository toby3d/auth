package memory

import (
	"context"
	"sync"

	"source.toby3d.me/toby3d/auth/internal/domain"
	"source.toby3d.me/toby3d/auth/internal/profile"
)

type memoryProfileRepository struct {
	mutex    *sync.RWMutex
	profiles map[string]domain.Profile
}

func NewMemoryProfileRepository() profile.Repository {
	return &memoryProfileRepository{
		mutex:    new(sync.RWMutex),
		profiles: make(map[string]domain.Profile),
	}
}

func (repo *memoryProfileRepository) Create(_ context.Context, me domain.Me, p domain.Profile) error {
	repo.mutex.Lock()
	defer repo.mutex.Unlock()

	repo.profiles[me.String()] = p

	return nil
}

func (repo *memoryProfileRepository) Get(_ context.Context, me domain.Me) (*domain.Profile, error) {
	repo.mutex.RLock()
	defer repo.mutex.RUnlock()

	if p, ok := repo.profiles[me.String()]; ok {
		return &p, nil
	}

	return nil, profile.ErrNotExist
}
