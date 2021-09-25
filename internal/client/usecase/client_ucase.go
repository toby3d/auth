package usecase

import (
	"context"

	"github.com/pkg/errors"
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
	c, err := useCase.clients.Get(ctx, string(clientID))
	if err != nil {
		return nil, errors.Wrap(err, "failed to get client information")
	}

	return c, nil
}
