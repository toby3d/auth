package token

import (
	"context"

	"source.toby3d.me/website/oauth/internal/model"
)

type Repository interface {
	Create(ctx context.Context, token *model.Token) error
	Get(ctx context.Context, token string) (*model.Token, error)
	Delete(ctx context.Context, token string) error
}
