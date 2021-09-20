package token

import (
	"context"

	"source.toby3d.me/website/oauth/internal/model"
)

type UseCase interface {
	Verify(ctx context.Context, token string) (*model.Token, error)
	Revoke(ctx context.Context, token string) error
}
