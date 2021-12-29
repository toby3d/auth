package domain

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type Ticket struct {
	// A random string that can be redeemed for an access token.
	Ticket string

	// The access token will work at this URL.
	Resource *URL

	// The access token should be used when acting on behalf of this URL.
	Subject *Me
}

func TestTicket(tb testing.TB) *Ticket {
	tb.Helper()

	subject, err := NewMe("https://bob.example.org/")
	require.NoError(tb, err)

	resource, err := NewURL("https://alice.example.com/private/")
	require.NoError(tb, err)

	return &Ticket{
		Ticket:   "32985723984723985792834",
		Resource: resource,
		Subject:  subject,
	}
}
