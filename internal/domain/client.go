package domain

import (
	"testing"
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

	redirects := make([]*URL, 0)
	for _, redirect := range []string{
		"https://app.example.net/redirect",
		"https://app.example.com/redirect",
	} {
		redirects = append(redirects, TestURL(tb, redirect))
	}

	return &Client{
		ID:          TestClientID(tb),
		Name:        []string{"Example App"},
		URL:         []*URL{TestURL(tb, "https://app.example.com/")},
		Logo:        []*URL{TestURL(tb, "https://app.example.com/logo.png")},
		RedirectURI: redirects,
	}
}
