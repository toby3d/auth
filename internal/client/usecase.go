package client

import (
	"context"

	"source.toby3d.me/website/oauth/internal/domain"
)

type UseCase interface {
	Discovery(ctx context.Context, clientID string) (*domain.Client, error)
}
