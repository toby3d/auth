package domain

import (
	"net"
	"net/url"
	"strings"
	"testing"
)

// Client describes the client requesting data about the user.
type Client struct {
	ID          *ClientID
	Logo        []*url.URL
	RedirectURI []*url.URL
	URL         []*url.URL
	Name        []string
}

// NewClient creates a new empty Client with provided ClientID, if any.
func NewClient(cid *ClientID) *Client {
	return &Client{
		ID:          cid,
		Logo:        make([]*url.URL, 0),
		RedirectURI: make([]*url.URL, 0),
		URL:         make([]*url.URL, 0),
		Name:        make([]string, 0),
	}
}

// TestClient returns valid random generated client for tests.
func TestClient(tb testing.TB) *Client {
	tb.Helper()

	return &Client{
		ID:   TestClientID(tb),
		Name: []string{"Example App"},
		URL:  []*url.URL{{Scheme: "https", Host: "app.example.com", Path: "/"}},
		Logo: []*url.URL{{Scheme: "https", Host: "app.example.com", Path: "/logo.png"}},
		RedirectURI: []*url.URL{
			{Scheme: "https", Host: "app.example.com", Path: "/redirect"},
			{Scheme: "https", Host: "app.example.net", Path: "/redirect"},
		},
	}
}

// ValidateRedirectURI validates RedirectURI from request to ClientID or
// registered set of client RedirectURI.
//
// If the URL scheme, host or port of the redirect_uri in the request do not
// match that of the client_id, then the authorization endpoint SHOULD verify
// that the requested redirect_uri matches one of the redirect URLs published by
// the client, and SHOULD block the request from proceeding if not.
func (c *Client) ValidateRedirectURI(redirectURI *url.URL) bool {
	if redirectURI == nil {
		return false
	}

	rHost, rPort, err := net.SplitHostPort(redirectURI.Host)
	if err != nil {
		rHost = redirectURI.Hostname()
	}

	cHost, cPort, err := net.SplitHostPort(c.ID.clientID.Host)
	if err != nil {
		cHost = c.ID.clientID.Hostname()
	}

	if strings.EqualFold(redirectURI.Scheme, c.ID.clientID.Scheme) &&
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

// GetName safe returns first name, if any.
func (c Client) GetName() string {
	if len(c.Name) == 0 {
		return ""
	}

	return c.Name[0]
}

// GetURL safe returns first URL, if any.
func (c Client) GetURL() *url.URL {
	if len(c.URL) == 0 {
		return nil
	}

	return c.URL[0]
}

// GetLogo safe returns first logo, if any.
func (c Client) GetLogo() *url.URL {
	if len(c.Logo) == 0 {
		return nil
	}

	return c.Logo[0]
}
