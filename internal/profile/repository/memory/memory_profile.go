package memory

import (
	"context"
	"fmt"
	"path"
	"sync"

	"source.toby3d.me/website/indieauth/internal/domain"
	"source.toby3d.me/website/indieauth/internal/profile"
)

type memoryProfileRepository struct {
	store *sync.Map
}

const (
	ErrPrefix         string = "memory"
	DefaultPathPrefix string = "profiles"
)

func NewMemoryProfileRepository(store *sync.Map) profile.Repository {
	return &memoryProfileRepository{
		store: store,
	}
}

func (repo *memoryProfileRepository) Get(_ context.Context, me *domain.Me) (*domain.Profile, error) {
	src, ok := repo.store.Load(path.Join(DefaultPathPrefix, me.String()))
	if !ok {
		return nil, fmt.Errorf("%s: cannot find profile in store: %w", ErrPrefix, profile.ErrNotExist)
	}

	result, ok := src.(*domain.Profile)
	if !ok {
		return nil, fmt.Errorf("%s: cannot decode profile from store: %w", ErrPrefix, profile.ErrNotExist)
	}

	return result, nil
}
