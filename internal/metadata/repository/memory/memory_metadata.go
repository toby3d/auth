package memory

import (
	"context"
	"path"
	"sync"

	"source.toby3d.me/website/indieauth/internal/domain"
	"source.toby3d.me/website/indieauth/internal/metadata"
)

type memoryMetadataRepository struct {
	store *sync.Map
}

const DefaultPathPrefix = "metadata"

func NewMemoryMetadataRepository(store *sync.Map) metadata.Repository {
	return &memoryMetadataRepository{
		store: store,
	}
}

func (repo *memoryMetadataRepository) Get(ctx context.Context, me *domain.Me) (*domain.Metadata, error) {
	src, ok := repo.store.Load(path.Join(DefaultPathPrefix, me.String()))
	if !ok {
		return nil, metadata.ErrNotExist
	}

	result, ok := src.(*domain.Metadata)
	if !ok {
		return nil, metadata.ErrNotExist
	}

	return result, nil
}
