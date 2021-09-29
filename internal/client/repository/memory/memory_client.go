package memory

import (
	"context"
	"sync"

	"source.toby3d.me/website/oauth/internal/client"
	"source.toby3d.me/website/oauth/internal/domain"
)

type memoryClientRepository struct {
	clients *sync.Map
}

func NewMemoryClientRepository(clients *sync.Map) client.Repository {
	return &memoryClientRepository{
		clients: clients,
	}
}

func (repo *memoryClientRepository) Get(ctx context.Context, id string) (*domain.Client, error) {
	src, ok := repo.clients.Load(id)
	if !ok {
		return nil, nil
	}

	c, ok := src.(*domain.Client)
	if !ok {
		return nil, nil
	}

	return c, nil
}
