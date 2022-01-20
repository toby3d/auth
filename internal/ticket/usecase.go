package ticket

import (
	"context"

	"source.toby3d.me/website/indieauth/internal/domain"
)

type UseCase interface {
	Generate(ctx context.Context, ticket *domain.Ticket) error

	// Exchange transform received ticket into access token.
	Exchange(ctx context.Context, ticket *domain.Ticket) (*domain.Token, error)
}
