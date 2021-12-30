package memory

import (
	"context"
	"path"
	"sync"

	"source.toby3d.me/website/oauth/internal/domain"
	"source.toby3d.me/website/oauth/internal/ticket"
)

type memoryTicketRepository struct {
	store *sync.Map
}

const DefaultPathPrefix string = "tickets"

func NewMemoryTicketRepository(store *sync.Map) ticket.Repository {
	return &memoryTicketRepository{
		store: store,
	}
}

func (repo *memoryTicketRepository) Get(_ context.Context, resource *domain.URL) (*domain.URL, error) {
	src, ok := repo.store.Load(path.Join(DefaultPathPrefix, resource.String()))
	if !ok {
		return nil, ticket.ErrNotExist
	}

	result, ok := src.(*domain.URL)
	if !ok {
		return nil, ticket.ErrNotExist
	}

	return result, nil
}
