package metadata

import (
	"context"

	"source.toby3d.me/toby3d/auth/internal/domain"
)

type Repository interface {
	Get(ctx context.Context, me *domain.Me) (*domain.Metadata, error)
}

var ErrNotExist error = domain.NewError(
	domain.ErrorCodeInvalidClient,
	"not found 'indieauth-metadata' endpoint on provided me URL",
	"https://indieauth.net/source/#discovery-0",
)
