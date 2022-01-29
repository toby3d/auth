package domain

import (
	"strings"
	"testing"
)

// Email represent email identifier.
type Email struct {
	user string
	host string
}

var ErrEmailInvalid error = NewError(ErrorCodeInvalidRequest, "cannot parse email", "")

// ParseEmail parse strings to email identifier.
func ParseEmail(src string) (*Email, error) {
	parts := strings.Split(strings.TrimPrefix(src, "mailto:"), "@")
	if len(parts) != 2 { //nolint: gomnd
		return nil, ErrEmailInvalid
	}

	return &Email{
		user: parts[0],
		host: parts[1],
	}, nil
}

// TestEmail returns valid random generated email identifier.
func TestEmail(tb testing.TB) *Email {
	tb.Helper()

	return &Email{
		user: "user",
		host: "example.com",
	}
}

// String returns string representation of email identifier.
func (e Email) String() string {
	return e.user + "@" + e.host
}
