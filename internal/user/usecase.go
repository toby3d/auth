package user

import (
	"context"

	"source.toby3d.me/website/indieauth/internal/domain"
)

type UseCase interface {
	// Fetch discovery all available endpoints and Profile info on Me URL.
	Fetch(ctx context.Context, me *domain.Me) (*domain.User, error)
}
