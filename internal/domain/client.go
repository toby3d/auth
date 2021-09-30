package domain

import "testing"

type Client struct {
	ID          string
	Logo        string
	Name        string
	RedirectURI []string
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
		ID:   "http://app.example.com/",
		Name: "Example App",
		Logo: "http://app.example.com/logo.png",
		URL:  "http://app.example.com/",
		RedirectURI: []string{
			"http://app.example.com/redirect",
			"http://app.example.com/redirect",
		},
	}
}
