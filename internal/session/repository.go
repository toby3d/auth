package session

import (
	"context"
	"errors"

	"source.toby3d.me/website/indieauth/internal/domain"
)

type Repository interface {
	Create(ctx context.Context, session *domain.Session) error
	GetAndDelete(ctx context.Context, code string) (*domain.Session, error)
	GC()
}

var ErrNotExist = errors.New("session not exist")
