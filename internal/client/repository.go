package client

import (
	"context"

	"source.toby3d.me/toby3d/auth/internal/domain"
)

type Repository interface {
	Get(ctx context.Context, id *domain.ClientID) (*domain.Client, error)
}

var ErrNotExist error = domain.NewError(
	domain.ErrorCodeInvalidClient,
	"client with the specified ID does not exist",
	"",
)
