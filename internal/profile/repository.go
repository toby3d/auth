package profile

import (
	"context"

	"source.toby3d.me/toby3d/auth/internal/domain"
)

type Repository interface {
	Create(ctx context.Context, me domain.Me, profile domain.Profile) error
	Get(ctx context.Context, me domain.Me) (*domain.Profile, error)
}

var ErrNotExist error = domain.NewError(
	domain.ErrorCodeServerError,
	"no profile data for the provided Me",
	"https://indieweb.org/h-card",
)
