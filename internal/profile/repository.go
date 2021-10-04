package profile

import (
	"context"

	"golang.org/x/oauth2"

	"source.toby3d.me/website/oauth/internal/domain"
)

type Repository interface {
	Get(ctx context.Context, token oauth2.Token) (*domain.Profile, error)
}
