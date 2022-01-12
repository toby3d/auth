package domain

import (
	"bytes"
	"net"
	"strings"
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

// ValidateRedirectURI validates RedirectURI from request to ClientID or
// registered set of client RedirectURI.
//
// If the URL scheme, host or port of the redirect_uri in the request do not
// match that of the client_id, then the authorization endpoint SHOULD verify
// that the requested redirect_uri matches one of the redirect URLs published by
// the client, and SHOULD block the request from proceeding if not.
func (c *Client) ValidateRedirectURI(redirectURI *URL) bool {
	if redirectURI == nil {
		return false
	}

	rHost, rPort, err := net.SplitHostPort(string(redirectURI.Host()))
	if err != nil {
		rHost = string(redirectURI.Host())
	}

	cHost, cPort, err := net.SplitHostPort(string(c.ID.clientID.Host()))
	if err != nil {
		cHost = string(c.ID.clientID.Host())
	}

	if bytes.EqualFold(redirectURI.Scheme(), c.ID.clientID.Scheme()) &&
		strings.EqualFold(rHost, cHost) &&
		strings.EqualFold(rPort, cPort) {
		return true
	}

	for i := range c.RedirectURI {
		if redirectURI.String() != c.RedirectURI[i].String() {
			continue
		}

		return true
	}

	return false
}

func (c Client) GetName() string {
	if len(c.Name) < 1 {
		return ""
	}

	return c.Name[0]
}

func (c Client) GetURL() *URL {
	if len(c.URL) < 1 {
		return nil
	}

	return c.URL[0]
}

func (c Client) GetLogo() *URL {
	if len(c.Logo) < 1 {
		return nil
	}

	return c.Logo[0]
}
