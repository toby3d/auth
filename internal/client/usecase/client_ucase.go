package usecase

import (
	"context"
	"fmt"

	"source.toby3d.me/website/oauth/internal/client"
	"source.toby3d.me/website/oauth/internal/domain"
)

type clientUseCase struct {
	repo client.Repository
}

func NewClientUseCase(repo client.Repository) client.UseCase {
	return &clientUseCase{
		repo: repo,
	}
}

func (useCase *clientUseCase) Discovery(ctx context.Context, id *domain.ClientID) (*domain.Client, error) {
	c, err := useCase.repo.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("cannot discovery client by id: %w", err)
	}

	return c, nil
}
