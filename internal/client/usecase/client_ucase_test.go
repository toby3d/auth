package usecase_test

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	repository "source.toby3d.me/website/oauth/internal/client/repository/memory"
	"source.toby3d.me/website/oauth/internal/client/usecase"
	"source.toby3d.me/website/oauth/internal/model"
)

func TestDiscovery(t *testing.T) {
	t.Parallel()

	require := require.New(t)
	assert := assert.New(t)
	store := new(sync.Map)
	client := &model.Client{
		ID:   "http://127.0.0.1:2368/",
		Name: "Example App",
		Logo: "http://127.0.0.1:2368/logo.png",
		URL:  "http://127.0.0.1:2368/",
		RedirectURI: []model.URL{
			"https://app.example.com/redirect",
			"http://127.0.0.1:2368/redirect",
		},
	}

	store.Store(string(client.ID), client)

	result, err := usecase.NewClientUseCase(repository.NewMemoryClientRepository(store)).Discovery(context.TODO(),
		client.ID)
	require.NoError(err)
	assert.Equal(client, result)
}
