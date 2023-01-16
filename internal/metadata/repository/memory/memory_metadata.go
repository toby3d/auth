package memory

import (
	"context"
	"net/url"
	"sync"

	"source.toby3d.me/toby3d/auth/internal/domain"
	"source.toby3d.me/toby3d/auth/internal/metadata"
)

type memoryMetadataRepository struct {
	mutex    *sync.RWMutex
	metadata map[string]domain.Metadata
}

const DefaultPathPrefix = "metadata"

func NewMemoryMetadataRepository() metadata.Repository {
	return &memoryMetadataRepository{
		mutex:    new(sync.RWMutex),
		metadata: make(map[string]domain.Metadata),
	}
}

func (repo *memoryMetadataRepository) Create(ctx context.Context, u *url.URL, metadata domain.Metadata) error {
	repo.mutex.Lock()
	defer repo.mutex.Unlock()

	repo.metadata[u.String()] = metadata

	return nil
}

func (repo *memoryMetadataRepository) Get(ctx context.Context, u *url.URL) (*domain.Metadata, error) {
	repo.mutex.RLock()
	defer repo.mutex.RUnlock()

	if out, ok := repo.metadata[u.String()]; ok {
		return &out, nil
	}

	return nil, metadata.ErrNotExist
}
