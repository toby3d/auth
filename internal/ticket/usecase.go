package ticket

import (
	"context"

	"source.toby3d.me/website/oauth/internal/domain"
)

type UseCase interface {
	// Redeem transform received ticket into access token.
	Redeem(ctx context.Context, endpoint *domain.URL, ticket string) (*domain.Token, error)
}
