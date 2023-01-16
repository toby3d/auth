package user

import (
	"context"

	"source.toby3d.me/toby3d/auth/internal/domain"
)

type Repository interface {
	Create(ctx context.Context, user domain.User) error
	Get(ctx context.Context, me domain.Me) (*domain.User, error)
}

var ErrNotExist error = domain.NewError(domain.ErrorCodeServerError, "user not exist", "")
