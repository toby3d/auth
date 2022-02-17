package profile

import (
	"context"

	"source.toby3d.me/website/indieauth/internal/domain"
)

type UseCase interface {
	Fetch(ctx context.Context, me *domain.Me) (*domain.Profile, error)
}

var ErrScopeRequired error = domain.NewError(
	domain.ErrorCodeInsufficientScope,
	"token with 'profile' scopes is required to view profile data",
	"https://indieauth.net/source/#user-information",
)
