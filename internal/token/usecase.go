package token

import (
	"context"
	"errors"

	"source.toby3d.me/website/oauth/internal/domain"
)

type UseCase interface {
	Verify(ctx context.Context, accessToken string) (*domain.Token, error)
	Revoke(ctx context.Context, accessToken string) error
}

var ErrRevoke error = errors.New("this token has been revoked")
