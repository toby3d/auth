package auth

import (
	"context"

	"gitlab.com/toby3d/indieauth/internal/domain"
)

type UseCase interface {
	Discovery(ctx context.Context, clientId string) (*domain.Client, error)
	Approve(ctx context.Context, login *domain.Login) (string, error)
	Exchange(ctx context.Context, req *domain.ExchangeRequest) (string, error)
}
