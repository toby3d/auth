package client

import (
	"context"

	"source.toby3d.me/website/oauth/internal/model"
)

type UseCase interface {
	Discovery(ctx context.Context, clientID model.URL) (*model.Client, error)
}
