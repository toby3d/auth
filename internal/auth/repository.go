package auth

import (
	"context"

	"gitlab.com/toby3d/indieauth/internal/domain"
)

type Repository interface {
	Create(ctx context.Context, login *domain.Login) error
	Get(ctx context.Context, code string) (*domain.Login, error)
	Delete(ctx context.Context, code string) error
}
