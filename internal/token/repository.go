package token

import (
	"context"

	"source.toby3d.me/website/indieauth/internal/domain"
)

type Repository interface {
	Get(ctx context.Context, accessToken string) (*domain.Token, error)
	Create(ctx context.Context, accessToken *domain.Token) error
}

var (
	ErrExist    error = domain.NewError(domain.ErrorCodeServerError, "token already exist", "")
	ErrNotExist error = domain.NewError(domain.ErrorCodeServerError, "token not exist", "")
)
