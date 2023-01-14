package metadata

import (
	"context"
	"net/url"

	"source.toby3d.me/toby3d/auth/internal/domain"
)

type Repository interface {
	Create(_ context.Context, _ *url.URL, _ domain.Metadata) error
	Get(_ context.Context, u *url.URL) (*domain.Metadata, error)
}

var ErrNotExist error = domain.NewError(
	domain.ErrorCodeInvalidClient,
	"not found 'indieauth-metadata' endpoint on provided me URL",
	"https://indieauth.net/source/#discovery-0",
)
