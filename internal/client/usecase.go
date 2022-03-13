package client

import (
	"context"

	"source.toby3d.me/toby3d/auth/internal/domain"
)

type UseCase interface {
	// Discovery returns client public information bu ClientID URL.
	Discovery(ctx context.Context, id *domain.ClientID) (*domain.Client, error)
}

var ErrInvalidMe error = domain.NewError(
	domain.ErrorCodeInvalidRequest,
	"cannot fetch client endpoints on provided me",
	"",
)
