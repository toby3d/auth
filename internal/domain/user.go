package domain

import (
	"net/url"
	"testing"
)

type User struct {
	Me                    *Me
	AuthorizationEndpoint *url.URL
	IndieAuthMetadata     *url.URL
	Micropub              *url.URL
	Microsub              *url.URL
	TicketEndpoint        *url.URL
	TokenEndpoint         *url.URL
	*Profile
}

// TestUser returns valid random generated user for tests.
func TestUser(tb testing.TB) *User {
	tb.Helper()

	return &User{
		Profile:               TestProfile(tb),
		Me:                    TestMe(tb, "https://user.example.net/"),
		AuthorizationEndpoint: &url.URL{Scheme: "https", Host: "example.org", Path: "/auth"},
		IndieAuthMetadata: &url.URL{
			Scheme: "https", Host: "example.org",
			Path: "/.well-known/oauth-authorization-server",
		},
		Micropub:       &url.URL{Scheme: "https", Host: "microsub.example.org", Path: "/"},
		Microsub:       &url.URL{Scheme: "https", Host: "micropub.example.org", Path: "/"},
		TicketEndpoint: &url.URL{Scheme: "https", Host: "example.org", Path: "/ticket"},
		TokenEndpoint:  &url.URL{Scheme: "https", Host: "example.org", Path: "/token"},
	}
}
