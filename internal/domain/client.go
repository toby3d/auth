package domain

import "testing"

type Client struct {
	RedirectURI []string
	ID          string
	Logo        string
	Name        string
	URL         string
}

func NewClient() *Client {
	c := new(Client)
	c.RedirectURI = make([]string, 0)

	return c
}

func TestClient(tb testing.TB) *Client {
	tb.Helper()

	return &Client{
		ID:   "https://app.example.com/",
		Name: "Example App",
		Logo: "https://app.example.com/logo.png",
		URL:  "https://app.example.com/",
		RedirectURI: []string{
			"https://app.example.net/redirect",
			"https://app.example.com/redirect",
		},
	}
}
