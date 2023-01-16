package session

import (
	"context"

	"source.toby3d.me/toby3d/auth/internal/domain"
)

type Repository interface {
	Get(ctx context.Context, code string) (*domain.Session, error)
	Create(ctx context.Context, session domain.Session) error
	GetAndDelete(ctx context.Context, code string) (*domain.Session, error)
	GC()
}

var ErrNotExist error = domain.NewError(domain.ErrorCodeServerError, "session with this code not exist", "")
