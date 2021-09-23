package token

import (
	"context"

	"source.toby3d.me/website/oauth/internal/model"
)

type UseCase interface {
	Verify(ctx context.Context, accessToken string) (*model.Token, error)
	Revoke(ctx context.Context, accessToken string) error
}
