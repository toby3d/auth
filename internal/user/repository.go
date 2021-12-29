package user

import (
	"context"
	"errors"

	"source.toby3d.me/website/oauth/internal/domain"
)

type Repository interface {
	Get(ctx context.Context, me *domain.Me) (*domain.User, error)
}

var ErrNotExist = errors.New("user not exists")
