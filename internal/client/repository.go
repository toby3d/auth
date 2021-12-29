package client

import (
	"context"
	"errors"

	"source.toby3d.me/website/oauth/internal/domain"
)

type Repository interface {
	Get(ctx context.Context, id *domain.ClientID) (*domain.Client, error)
}

var ErrNotExist = errors.New("client with the specified ID does not exist")
