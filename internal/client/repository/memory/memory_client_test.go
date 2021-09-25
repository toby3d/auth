package memory_test

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"source.toby3d.me/website/oauth/internal/client/repository/memory"
	"source.toby3d.me/website/oauth/internal/model"
)

func TestGet(t *testing.T) {
	t.Parallel()

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

	result, err := memory.NewMemoryClientRepository(store).Get(context.TODO(), string(client.ID))
	require.NoError(t, err)
	assert.Equal(t, client, result)
}
