package memory

import (
	"context"
	"fmt"
	"path"
	"sync"
	"time"

	"source.toby3d.me/website/indieauth/internal/domain"
	"source.toby3d.me/website/indieauth/internal/session"
)

type (
	Session struct {
		CreatedAt time.Time
		*domain.Session
	}

	memorySessionRepository struct {
		store  *sync.Map
		config *domain.Config
	}
)

const DefaultPathPrefix string = "sessions"

func NewMemorySessionRepository(config *domain.Config, store *sync.Map) session.Repository {
	return &memorySessionRepository{
		config: config,
		store:  store,
	}
}

func (repo *memorySessionRepository) Create(_ context.Context, state *domain.Session) error {
	repo.store.Store(path.Join(DefaultPathPrefix, state.Code), &Session{
		CreatedAt: time.Now().UTC(),
		Session:   state,
	})

	return nil
}

func (repo *memorySessionRepository) Get(_ context.Context, code string) (*domain.Session, error) {
	src, ok := repo.store.Load(path.Join(DefaultPathPrefix, code))
	if !ok {
		return nil, fmt.Errorf("cannot find session in store: %w", session.ErrNotExist)
	}

	result, ok := src.(*Session)
	if !ok {
		return nil, fmt.Errorf("cannot decode session in store: %w", session.ErrNotExist)
	}

	return result.Session, nil
}

func (repo *memorySessionRepository) GetAndDelete(_ context.Context, code string) (*domain.Session, error) {
	src, ok := repo.store.LoadAndDelete(path.Join(DefaultPathPrefix, code))
	if !ok {
		return nil, fmt.Errorf("cannot find session in store: %w", session.ErrNotExist)
	}

	result, ok := src.(*Session)
	if !ok {
		return nil, fmt.Errorf("cannot decode session in store: %w", session.ErrNotExist)
	}

	return result.Session, nil
}

func (repo *memorySessionRepository) GC() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for ts := range ticker.C {
		ts := ts

		repo.store.Range(func(key, value interface{}) bool {
			k, ok := key.(string)
			if !ok {
				return false
			}

			matched, err := path.Match(DefaultPathPrefix+"/*", k)
			if err != nil || !matched {
				return false
			}

			val, ok := value.(*Session)
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
