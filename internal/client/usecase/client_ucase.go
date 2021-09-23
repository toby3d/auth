package usecase

import (
	"context"

	"source.toby3d.me/website/oauth/internal/client"
	"source.toby3d.me/website/oauth/internal/model"
)

type clientUseCase struct {
	clients client.Repository
}

func NewClientUseCase(clients client.Repository) client.UseCase {
	return &clientUseCase{
		clients: clients,
	}
}

func (useCase *clientUseCase) Discovery(ctx context.Context, clientID model.URL) (*model.Client, error) {
	return useCase.clients.Get(ctx, string(clientID))
}
