package auth

import (
	"context"

	"gitlab.com/toby3d/indieauth/internal/model"
)

type UseCase interface {
	Discovery(ctx context.Context, clientId string) (*model.Client, error)
	Approve(ctx context.Context, login *model.Login) (string, error)
	Exchange(ctx context.Context, req *model.ExchangeRequest) (string, error)
}
