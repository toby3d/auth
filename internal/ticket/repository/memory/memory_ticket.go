package memory

import (
	"context"
	"fmt"
	"path"
	"sync"
	"time"

	"source.toby3d.me/website/indieauth/internal/domain"
	"source.toby3d.me/website/indieauth/internal/ticket"
)

type (
	Ticket struct {
		CreatedAt time.Time
		*domain.Ticket
	}

	memoryTicketRepository struct {
		config *domain.Config
		store  *sync.Map
	}
)

const DefaultPathPrefix string = "tickets"

func NewMemoryTicketRepository(store *sync.Map, config *domain.Config) ticket.Repository {
	return &memoryTicketRepository{
		config: config,
		store:  store,
	}
}

func (repo *memoryTicketRepository) Create(_ context.Context, t *domain.Ticket) error {
	repo.store.Store(path.Join(DefaultPathPrefix, t.Ticket), &Ticket{
		CreatedAt: time.Now().UTC(),
		Ticket:    t,
	})

	return nil
}

func (repo *memoryTicketRepository) GetAndDelete(_ context.Context, t string) (*domain.Ticket, error) {
	src, ok := repo.store.LoadAndDelete(path.Join(DefaultPathPrefix, t))
	if !ok {
		return nil, fmt.Errorf("cannot find ticket in store: %w", ticket.ErrNotExist)
	}

	result, ok := src.(*Ticket)
	if !ok {
		return nil, fmt.Errorf("cannot decode ticket in store: %w", ticket.ErrNotExist)
	}

	return result.Ticket, nil
}

func (repo *memoryTicketRepository) GC() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for ts := range ticker.C {
		ts := ts.UTC()

		repo.store.Range(func(key, value interface{}) bool {
			k, ok := key.(string)
			if !ok {
				return false
			}

			matched, err := path.Match(DefaultPathPrefix+"/*", k)
			if err != nil || !matched {
				return false
			}

			val, ok := value.(*Ticket)
			if !ok {
				return false
			}

			if val.CreatedAt.Add(repo.config.Code.Expiry).After(ts) {
				return false
			}

			repo.store.Delete(key)

			return false
		})
	}
}
