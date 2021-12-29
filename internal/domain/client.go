package domain

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Client describes the client requesting data about the user.
type Client struct {
	ID          *ClientID
	Logo        []*URL
	RedirectURI []*URL
	URL         []*URL
	Name        []string
}

// TestClient returns a valid Client with the generated test data filled in.
func TestClient(tb testing.TB) *Client {
	tb.Helper()

	url, err := NewURL("https://app.example.com/")
	require.NoError(tb, err)

	logo, err := NewURL("https://app.example.com/logo.png")
	require.NoError(tb, err)

	redirects := make([]*URL, 0)

	for _, redirect := range []string{
		"https://app.example.net/redirect",
		"https://app.example.com/redirect",
	} {
		u, err := NewURL(redirect)
		require.NoError(tb, err)

		redirects = append(redirects, u)
	}

	return &Client{
		ID:          TestClientID(tb),
		Name:        []string{"Example App"},
		URL:         []*URL{url},
		Logo:        []*URL{logo},
		RedirectURI: redirects,
	}
}
