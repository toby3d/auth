package domain

import (
	"strings"
	"testing"
)

// Email represent email identifier.
type Email struct {
	user       string
	host       string
	subAddress string
}

var ErrEmailInvalid error = NewError(ErrorCodeInvalidRequest, "cannot parse email", "")

// ParseEmail parse strings to email identifier.
func ParseEmail(src string) (*Email, error) {
	parts := strings.Split(strings.TrimPrefix(src, "mailto:"), "@")
	if len(parts) != 2 { //nolint: gomnd
		return nil, ErrEmailInvalid
	}

	result := &Email{
		user:       parts[0],
		host:       parts[1],
		subAddress: "",
	}

	if userParts := strings.SplitN(parts[0], `+`, 2); len(userParts) > 1 {
		result.user = userParts[0]
		result.subAddress = userParts[1]
	}

	return result, nil
}

// TestEmail returns valid random generated email identifier.
func TestEmail(tb testing.TB) *Email {
	tb.Helper()

	return &Email{
		user:       "user",
		subAddress: "",
		host:       "example.com",
	}
}

// String returns string representation of email identifier.
func (e Email) String() string {
	if e.subAddress == "" {
		return e.user + "@" + e.host
	}

	return e.user + "+" + e.subAddress + "@" + e.host
}
