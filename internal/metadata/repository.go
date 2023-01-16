package metadata

import (
	"context"
	"net/url"

	"source.toby3d.me/toby3d/auth/internal/domain"
)

type Repository interface {
	Create(ctx context.Context, u *url.URL, metadata domain.Metadata) error
	Get(ctx context.Context, u *url.URL) (*domain.Metadata, error)
}

var ErrNotExist error = domain.NewError(
	domain.ErrorCodeInvalidClient,
	"not found 'indieauth-metadata' endpoint on provided me URL",
	"https://indieauth.net/source/#discovery-0",
)
