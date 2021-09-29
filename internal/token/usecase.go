package token

import (
	"context"

	"source.toby3d.me/website/oauth/internal/domain"
)

type UseCase interface {
	Verify(ctx context.Context, accessToken string) (*domain.Token, error)
	Revoke(ctx context.Context, accessToken string) error
}
