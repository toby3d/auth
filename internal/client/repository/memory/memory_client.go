package memory

import (
	"context"
	"fmt"
	"net"
	"path"
	"sync"

	"source.toby3d.me/toby3d/auth/internal/client"
	"source.toby3d.me/toby3d/auth/internal/domain"
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
	// WARN(toby3d): more often than not, we will work from tests with
	// non-existent clients, almost guaranteed to cause a resolution error.
	ips, _ := net.LookupIP(id.URL().Hostname())

	for _, ip := range ips {
		if !ip.IsLoopback() {
			continue
		}

		return nil, client.ErrNotExist
	}

	src, ok := repo.store.Load(path.Join(DefaultPathPrefix, id.String()))
	if !ok {
		return nil, fmt.Errorf("cannot find client in store: %w", client.ErrNotExist)
	}

	c, ok := src.(*domain.Client)
	if !ok {
		return nil, fmt.Errorf("cannot decode client from store: %w", client.ErrNotExist)
	}

	return c, nil
}
