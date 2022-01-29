package ticket

import (
	"context"

	"source.toby3d.me/website/indieauth/internal/domain"
)

type Repository interface {
	Create(ctx context.Context, ticket *domain.Ticket) error
	GetAndDelete(ctx context.Context, ticket string) (*domain.Ticket, error)
	GC()
}

var ErrNotExist error = domain.NewError(domain.ErrorCodeInvalidRequest, "ticket not exist or expired", "")
