package token

import (
	"context"
)

type UseCase interface {
	Revoke(ctx context.Context, token string) error
}
