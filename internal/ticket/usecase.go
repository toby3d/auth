package ticket

import (
	"context"

	"source.toby3d.me/toby3d/auth/internal/domain"
)

type UseCase interface {
	Generate(ctx context.Context, ticket domain.Ticket) error

	// Redeem transform received ticket into access token.
	Redeem(ctx context.Context, ticket domain.Ticket) (*domain.Token, error)
	Exchange(ctx context.Context, ticket string) (*domain.Token, error)
}

var (
	ErrTicketEndpointNotExist error = domain.NewError(
		domain.ErrorCodeServerError, "ticket_endpoint not found on ticket resource", "",
	)
	ErrTokenEndpointNotExist error = domain.NewError(
		domain.ErrorCodeServerError, "token_endpoint not found on ticket resource", "",
	)
)
