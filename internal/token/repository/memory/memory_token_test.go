package memory_test

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"source.toby3d.me/website/oauth/internal/model"
	"source.toby3d.me/website/oauth/internal/random"
	"source.toby3d.me/website/oauth/internal/token"
	"source.toby3d.me/website/oauth/internal/token/repository/memory"
)

func TestGet(t *testing.T) {
	t.Parallel()

	store := new(sync.Map)
	accessToken := &model.Token{
		AccessToken: random.New().String(32),
		ClientID:    "https://app.example.com/",
		Me:          "https://toby3d.me/",
		Profile: &model.Profile{
			Name:  "Maxim Lebedev",
			URL:   "https://toby3d.me/",
			Photo: "https://toby3d.me/photo.jpg",
			Email: "hey@toby3d.me",
		},
		Scopes: []string{"read", "update", "delete"},
		Type:   "Bearer",
	}

	store.Store(accessToken.AccessToken, accessToken)

	result, err := memory.NewMemoryTokenRepository(store).Get(context.TODO(), accessToken.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, accessToken, result)
}

func TestCreate(t *testing.T) {
	t.Parallel()

	store := new(sync.Map)
	accessToken := &model.Token{
		AccessToken: random.New().String(32),
		ClientID:    "https://app.example.com/",
		Me:          "https://toby3d.me/",
		Profile: &model.Profile{
			Name:  "Maxim Lebedev",
			URL:   "https://toby3d.me/",
			Photo: "https://toby3d.me/photo.jpg",
			Email: "hey@toby3d.me",
		},
		Scopes: []string{"read", "update", "delete"},
		Type:   "Bearer",
	}

	repo := memory.NewMemoryTokenRepository(store)
	require.NoError(t, repo.Create(context.TODO(), accessToken))

	result, ok := store.Load(accessToken.AccessToken)
	assert.True(t, ok)
	assert.Equal(t, accessToken, result)

	assert.EqualError(t, repo.Create(context.TODO(), accessToken), token.ErrExist.Error())
}

func TestUpdate(t *testing.T) {
	t.Parallel()

	store := new(sync.Map)
	accessToken := &model.Token{
		AccessToken: random.New().String(32),
		ClientID:    "https://app.example.com/",
		Me:          "https://toby3d.me/",
		Profile: &model.Profile{
			Name:  "Maxim Lebedev",
			URL:   "https://toby3d.me/",
			Photo: "https://toby3d.me/photo.jpg",
			Email: "hey@toby3d.me",
		},
		Scopes: []string{"read", "update", "delete"},
		Type:   "Bearer",
	}

	store.Store(accessToken.AccessToken, accessToken)

	tokenCopy := *accessToken
	tokenCopy.ClientID = "https://client.example.com/"
	tokenCopy.Me = "https://toby3d.ru/"

	require.NoError(t, memory.NewMemoryTokenRepository(store).Update(context.TODO(), &tokenCopy))

	result, ok := store.Load(accessToken.AccessToken)
	assert.True(t, ok)
	assert.NotEqual(t, accessToken, result)
	assert.Equal(t, &tokenCopy, result)
}

func TestDelete(t *testing.T) {
	t.Parallel()

	store := new(sync.Map)
	accessToken := &model.Token{
		AccessToken: random.New().String(32),
		ClientID:    "https://app.example.com/",
		Me:          "https://toby3d.me/",
		Profile: &model.Profile{
			Name:  "Maxim Lebedev",
			URL:   "https://toby3d.me/",
			Photo: "https://toby3d.me/photo.jpg",
			Email: "hey@toby3d.me",
		},
		Scopes: []string{"read", "update", "delete"},
		Type:   "Bearer",
	}

	store.Store(accessToken.AccessToken, accessToken)

	require.NoError(t, memory.NewMemoryTokenRepository(store).Remove(context.TODO(), accessToken.AccessToken))

	result, ok := store.Load(accessToken.AccessToken)
	assert.False(t, ok)
	assert.Nil(t, result)
}