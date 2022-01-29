package domain

import (
	"strings"
	"testing"
)

type Email struct {
	user string
	host string
}

var ErrEmailInvalid error = Error{
	Code:        ErrorCodeInvalidRequest,
	Description: "cannot parse email",
}

func NewEmail(src string) (*Email, error) {
	parts := strings.Split(strings.TrimPrefix(src, "mailto:"), "@")
	if len(parts) != 2 { //nolint: gomnd
		return nil, ErrEmailInvalid
	}

	return &Email{
		user: parts[0],
		host: parts[1],
	}, nil
}

func TestEmail(tb testing.TB) *Email {
	tb.Helper()

	return &Email{
		user: "user",
		host: "example.com",
	}
}

func (e Email) String() string {
	return e.user + "@" + e.host
}
