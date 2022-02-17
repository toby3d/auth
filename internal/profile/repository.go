package profile

import (
	"context"

	"source.toby3d.me/website/indieauth/internal/domain"
)

type Repository interface {
	Get(ctx context.Context, me *domain.Me) (*domain.Profile, error)
}

var ErrNotExist error = domain.NewError(
	domain.ErrorCodeServerError,
	"no profile data for the provided Me",
	"https://indieweb.org/h-card",
)
