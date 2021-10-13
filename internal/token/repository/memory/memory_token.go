package memory

import (
	"context"
	"path"
	"sync"
	"time"

	"source.toby3d.me/website/oauth/internal/domain"
	"source.toby3d.me/website/oauth/internal/token"
)

type memoryTokenRepository struct {
	store *sync.Map
}

const DefaultPathPrefix string = "tokens"

func NewMemoryTokenRepository(store *sync.Map) token.Repository {
	return &memoryTokenRepository{
		store: store,
	}
}

func (repo *memoryTokenRepository) Create(ctx context.Context, accessToken *domain.Token) error {
	key := path.Join(DefaultPathPrefix, accessToken.AccessToken)

	if _, ok := repo.store.Load(key); ok {
		return token.ErrExist
	}

	repo.store.Store(key, accessToken.Expiry)

	return nil
}

func (repo *memoryTokenRepository) Get(ctx context.Context, accessToken string) (*domain.Token, error) {
	expiry, ok := repo.store.Load(path.Join(DefaultPathPrefix, accessToken))
	if !ok {
		return nil, nil
	}

	return &domain.Token{
		Expiry:      expiry.(time.Time),
		Scopes:      []string{},
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ClientID:    "",
		Me:          "",
	}, nil
}
