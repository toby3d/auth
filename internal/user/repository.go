package user

import (
	"context"

	"source.toby3d.me/website/indieauth/internal/domain"
)

type Repository interface {
	Get(ctx context.Context, me *domain.Me) (*domain.User, error)
}

var ErrNotExist error = domain.NewError(domain.ErrorCodeServerError, "user not exist", "")
