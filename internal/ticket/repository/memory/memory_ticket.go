package memory

import (
	"context"
	"sync"
	"time"

	"source.toby3d.me/toby3d/auth/internal/domain"
	"source.toby3d.me/toby3d/auth/internal/ticket"
)

type (
	Ticket struct {
		CreatedAt time.Time
		domain.Ticket
	}

	memoryTicketRepository struct {
		mutex   *sync.RWMutex
		tickets map[string]Ticket
		config  domain.Config
	}
)

func NewMemoryTicketRepository(config domain.Config) ticket.Repository {
	return &memoryTicketRepository{
		config:  config,
		mutex:   new(sync.RWMutex),
		tickets: make(map[string]Ticket),
	}
}

func (repo *memoryTicketRepository) Create(_ context.Context, t domain.Ticket) error {
	repo.mutex.Lock()
	defer repo.mutex.Unlock()

	repo.tickets[t.Ticket] = Ticket{
		CreatedAt: time.Now().UTC(),
		Ticket:    t,
	}

	return nil
}

func (repo *memoryTicketRepository) GetAndDelete(_ context.Context, t string) (*domain.Ticket, error) {
	repo.mutex.RLock()

	out, ok := repo.tickets[t]
	if !ok {
		repo.mutex.RUnlock()

		return nil, ticket.ErrNotExist
	}

	repo.mutex.RUnlock()
	repo.mutex.Lock()
	delete(repo.tickets, t)
	repo.mutex.Unlock()

	return &out.Ticket, nil
}

func (repo *memoryTicketRepository) GC() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for ts := range ticker.C {
		ts = ts.UTC()

		repo.mutex.RLock()

		for _, t := range repo.tickets {
			if t.CreatedAt.Add(repo.config.Code.Expiry).After(ts) {
				continue
			}

			repo.mutex.RUnlock()
			repo.mutex.Lock()
			delete(repo.tickets, t.Ticket.Ticket)
			repo.mutex.Unlock()
			repo.mutex.RLock()
		}

		repo.mutex.RUnlock()
	}
}
