package token

import (
	"context"

	"source.toby3d.me/toby3d/auth/internal/domain"
)

type Repository interface {
	Create(ctx context.Context, accessToken domain.Token) error
	Get(ctx context.Context, accessToken string) (*domain.Token, error)
}

var (
	ErrExist    error = domain.NewError(domain.ErrorCodeServerError, "token already exist", "")
	ErrNotExist error = domain.NewError(domain.ErrorCodeServerError, "token not exist", "")
)
