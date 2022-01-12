package session

import (
	"context"

	"source.toby3d.me/website/indieauth/internal/domain"
)

type UseCase interface {
	Exchange(ctx context.Context, code string) (*domain.Session, error)
}
