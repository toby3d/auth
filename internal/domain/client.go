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
		ID:   "http://127.0.0.1:2368/",
		Name: "Example App",
		Logo: "http://127.0.0.1:2368/logo.png",
		URL:  "http://127.0.0.1:2368/",
		RedirectURI: []string{
			"https://app.example.com/redirect",
			"http://127.0.0.1:2368/redirect",
		},
	}
}
