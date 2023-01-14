package memory

import (
	"context"
	"sync"

	"source.toby3d.me/toby3d/auth/internal/client"
	"source.toby3d.me/toby3d/auth/internal/domain"
)

type memoryClientRepository struct {
	mutex   *sync.RWMutex
	clients map[string]domain.Client
}

func NewMemoryClientRepository() client.Repository {
	return &memoryClientRepository{
		mutex:   new(sync.RWMutex),
		clients: make(map[string]domain.Client),
	}
}

func (repo memoryClientRepository) Create(ctx context.Context, client domain.Client) error {
	repo.mutex.RLock()
	defer repo.mutex.RUnlock()

	repo.clients[client.ID.String()] = client

	return nil
}

func (repo memoryClientRepository) Get(ctx context.Context, cid domain.ClientID) (*domain.Client, error) {
	repo.mutex.RLock()
	defer repo.mutex.RUnlock()

	if c, ok := repo.clients[cid.String()]; ok {
		return &c, nil
	}

	return nil, client.ErrNotExist
}
