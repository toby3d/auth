package client

import (
	"context"

	"source.toby3d.me/website/oauth/internal/domain"
)

type Repository interface {
	Get(ctx context.Context, id string) (*domain.Client, error)
}
