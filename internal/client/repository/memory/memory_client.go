package memory

import (
	"context"
	"path"
	"sync"

	"source.toby3d.me/website/oauth/internal/client"
	"source.toby3d.me/website/oauth/internal/domain"
)

type memoryClientRepository struct {
	store *sync.Map
}

const DefaultPathPrefix string = "clients"

func NewMemoryClientRepository(store *sync.Map) client.Repository {
	return &memoryClientRepository{
		store: store,
	}
}

func (repo *memoryClientRepository) Create(ctx context.Context, client *domain.Client) error {
	repo.store.Store(path.Join(DefaultPathPrefix, client.ID.String()), client)

	return nil
}

func (repo *memoryClientRepository) Get(ctx context.Context, id *domain.ClientID) (*domain.Client, error) {
	src, ok := repo.store.Load(path.Join(DefaultPathPrefix, id.String()))
	if !ok {
		return nil, client.ErrNotExist
	}

	c, ok := src.(*domain.Client)
	if !ok {
		return nil, client.ErrNotExist
	}

	return c, nil
}
