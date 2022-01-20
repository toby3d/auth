package ticket

import (
	"context"
	"errors"

	"source.toby3d.me/website/indieauth/internal/domain"
)

type Repository interface {
	Create(ctx context.Context, ticket *domain.Ticket) error
	GetAndDelete(ctx context.Context, ticket string) (*domain.Ticket, error)
	GC()
}

var ErrNotExist = errors.New("token_endpoint not found on resource URL")
