package client

import (
	"context"
	"errors"

	"source.toby3d.me/website/oauth/internal/domain"
)

type UseCase interface {
	// Discovery returns client public information bu ClientID URL.
	Discovery(ctx context.Context, id *domain.ClientID) (*domain.Client, error)
}

var ErrInvalidMe = errors.New("provided me is invalid")
