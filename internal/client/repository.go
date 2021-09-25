package client

import (
	"context"

	"source.toby3d.me/website/oauth/internal/model"
)

type Repository interface {
	Get(ctx context.Context, id string) (*model.Client, error)
}
