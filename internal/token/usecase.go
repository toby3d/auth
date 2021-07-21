package token

import (
	"context"

	"gitlab.com/toby3d/indieauth/internal/model"
)

type UseCase interface {
	Exchange(ctx context.Context, req *model.ExchangeRequest) (*model.Token, error)
	Verify(ctx context.Context, token string) (*model.Token, error)
	Revoke(ctx context.Context, token string) error
}
