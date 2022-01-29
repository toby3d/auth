package domain

import (
	"testing"
)

type User struct {
	Me                    *Me
	AuthorizationEndpoint *URL
	IndieAuthMetadata     *URL
	Micropub              *URL
	Microsub              *URL
	TicketEndpoint        *URL
	TokenEndpoint         *URL
	*Profile
}

// TestUser returns valid random generated user for tests.
func TestUser(tb testing.TB) *User {
	tb.Helper()

	return &User{
		Me:                    TestMe(tb, "https://user.example.net/"),
		Profile:               TestProfile(tb),
		AuthorizationEndpoint: TestURL(tb, "https://example.org/auth"),
		IndieAuthMetadata:     TestURL(tb, "https://example.org/.well-known/oauth-authorization-server"),
		Micropub:              TestURL(tb, "https://microsub.example.org/"),
		Microsub:              TestURL(tb, "https://micropub.example.org/"),
		TicketEndpoint:        TestURL(tb, "https://example.org/ticket"),
		TokenEndpoint:         TestURL(tb, "https://example.org/token"),
	}
}
