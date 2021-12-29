package memory

import (
	"context"
	"path"
	"sync"

	"source.toby3d.me/website/oauth/internal/domain"
	"source.toby3d.me/website/oauth/internal/user"
)

type memoryUserRepository struct {
	store *sync.Map
}

const DefaultPathPrefix string = "users"

func NewMemoryUserRepository(store *sync.Map) user.Repository {
	return &memoryUserRepository{
		store: store,
	}
}

func (repo *memoryUserRepository) Get(ctx context.Context, me *domain.Me) (*domain.User, error) {
	p, ok := repo.store.Load(path.Join(DefaultPathPrefix, me.String()))
	if !ok {
		return nil, user.ErrNotExist
	}

	result, ok := p.(*domain.User)
	if !ok {
		return nil, user.ErrNotExist
	}

	return result, nil
}
