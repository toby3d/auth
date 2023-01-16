package client

import (
	"context"

	"source.toby3d.me/toby3d/auth/internal/domain"
)

type Repository interface {
	Create(ctx context.Context, client domain.Client) error
	Get(ctx context.Context, cid domain.ClientID) (*domain.Client, error)
}

var ErrNotExist error = domain.NewError(
	domain.ErrorCodeInvalidClient,
	"client with the specified ID does not exist",
	"",
)
