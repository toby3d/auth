package usecase

import (
	"context"
	"fmt"

	"source.toby3d.me/toby3d/auth/internal/domain"
	"source.toby3d.me/toby3d/auth/internal/user"
)

type userUseCase struct {
	repo user.Repository
}

func NewUserUseCase(repo user.Repository) user.UseCase {
	return &userUseCase{
		repo: repo,
	}
}

func (useCase *userUseCase) Fetch(ctx context.Context, me *domain.Me) (*domain.User, error) {
	user, err := useCase.repo.Get(ctx, me)
	if err != nil {
		return nil, fmt.Errorf("cannot find user by me: %w", err)
	}

	return user, nil
}
