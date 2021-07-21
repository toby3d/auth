package auth

import (
	"context"

	"gitlab.com/toby3d/indieauth/internal/model"
)

type Repository interface {
	Create(ctx context.Context, login *model.Login) error
	Get(ctx context.Context, code string) (*model.Login, error)
	Delete(ctx context.Context, code string) error
}
