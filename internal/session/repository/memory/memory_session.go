package memory

import (
	"context"
	"fmt"
	"sync"
	"time"

	"source.toby3d.me/toby3d/auth/internal/domain"
	"source.toby3d.me/toby3d/auth/internal/session"
)

type (
	Session struct {
		CreatedAt time.Time
		domain.Session
	}

	memorySessionRepository struct {
		mutex    *sync.RWMutex
		sessions map[string]Session
		config   domain.Config
	}
)

func NewMemorySessionRepository(config domain.Config) session.Repository {
	return &memorySessionRepository{
		config:   config,
		mutex:    new(sync.RWMutex),
		sessions: make(map[string]Session),
	}
}

func (repo *memorySessionRepository) Create(_ context.Context, s domain.Session) error {
	repo.mutex.Lock()
	defer repo.mutex.Unlock()

	repo.sessions[s.Code] = Session{
		CreatedAt: time.Now().UTC(),
		Session:   s,
	}

	return nil
}

func (repo *memorySessionRepository) Get(_ context.Context, code string) (*domain.Session, error) {
	repo.mutex.Lock()
	defer repo.mutex.Unlock()

	if s, ok := repo.sessions[code]; ok {
		return &s.Session, nil
	}

	return nil, session.ErrNotExist
}

func (repo *memorySessionRepository) GetAndDelete(ctx context.Context, code string) (*domain.Session, error) {
	s, err := repo.Get(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("cannot get and delete session: %w", err)
	}

	repo.mutex.Lock()
	defer repo.mutex.Unlock()

	delete(repo.sessions, s.Code)

	return s, nil
}

func (repo *memorySessionRepository) GC() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for ts := range ticker.C {
		ts := ts

		repo.mutex.RLock()

		for code, s := range repo.sessions {
			if s.CreatedAt.Add(repo.config.Code.Expiry).After(ts) {
				continue
			}

			repo.mutex.RUnlock()
			repo.mutex.Lock()
			delete(repo.sessions, code)
			repo.mutex.Unlock()
			repo.mutex.RLock()
		}

		repo.mutex.RUnlock()
	}
}
