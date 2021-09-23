package memory

import (
	"context"
	"sync"

	"source.toby3d.me/website/oauth/internal/client"
	"source.toby3d.me/website/oauth/internal/model"
)

type memoryClientRepository struct {
	clients *sync.Map
}

func NewMemoryClientRepository(clients *sync.Map) client.Repository {
	return &memoryClientRepository{
		clients: clients,
	}
}

func (repo *memoryClientRepository) Get(ctx context.Context, id string) (*model.Client, error) {
	src, ok := repo.clients.Load(id)
	if !ok {
		return nil, nil
	}

	c, ok := src.(*model.Client)
	if !ok {
		return nil, nil
	}

	return c, nil
}
