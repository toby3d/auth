package token

import (
	"context"
	"errors"

	"source.toby3d.me/website/oauth/internal/domain"
)

type Repository interface {
	Get(ctx context.Context, accessToken string) (*domain.Token, error)
	Create(ctx context.Context, accessToken *domain.Token) error
}

var (
	ErrExist    = errors.New("token already exist")
	ErrNotExist = errors.New("token not exist")
)
