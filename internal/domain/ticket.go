package domain

import (
	"net/url"
	"testing"
)

type Ticket struct {
	// The access token will work at this URL.
	Resource *url.URL

	// The access token should be used when acting on behalf of this URL.
	Subject *Me

	// A random string that can be redeemed for an access token.
	Ticket string
}

// TestTicket returns valid random generated ticket for tests.
func TestTicket(tb testing.TB) *Ticket {
	tb.Helper()

	return &Ticket{
		Resource: &url.URL{Scheme: "https", Host: "alice.example.com", Path: "/private/"},
		Subject:  TestMe(tb, "https://bob.example.com/"),
		Ticket:   "32985723984723985792834",
	}
}
