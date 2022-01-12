package usecase

import (
	"context"
	"fmt"

	"source.toby3d.me/website/indieauth/internal/domain"
	"source.toby3d.me/website/indieauth/internal/session"
)

type sessionUseCase struct {
	sessions session.Repository
}

func NewSessionUseCase(sessions session.Repository) session.UseCase {
	return &sessionUseCase{
		sessions: sessions,
	}
}

func (useCase *sessionUseCase) Exchange(ctx context.Context, code string) (*domain.Session, error) {
	session, err := useCase.sessions.GetAndDelete(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("cannot find session in store: %w", err)
	}

	return session, nil
}
