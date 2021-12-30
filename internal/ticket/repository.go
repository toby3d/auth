package ticket

import (
	"context"
	"errors"

	"source.toby3d.me/website/oauth/internal/domain"
)

type Repository interface {
	// Get returns token endpoint founded by resource URL.
	Get(ctx context.Context, resource *domain.URL) (*domain.URL, error)
}

var ErrNotExist = errors.New("token_endpoint not found on resource URL")
