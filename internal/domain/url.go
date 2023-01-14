package domain

import (
	"fmt"
	"net/url"
	"strconv"
	"testing"

	"source.toby3d.me/toby3d/auth/internal/common"
)

// URL describe any valid HTTP URL.
type URL struct {
	*url.URL
}

// ParseURL parse string as URL.
func ParseURL(src string) (*URL, error) {
	u, err := url.Parse(src)
	if err != nil {
		return nil, fmt.Errorf("cannot parse URL: %w", err)
	}

	return &URL{URL: u}, nil
}

// MustParseURL parse string as URL or panic.
func MustParseURL(src string) *URL {
	uri, err := ParseURL(src)
	if err != nil {
		panic("MustParseURL: " + err.Error())
	}

	return uri
}

// TestURL returns URL of provided input for tests.
func TestURL(tb testing.TB, src string) *URL {
	tb.Helper()

	u, _ := url.Parse(src)

	return &URL{
		URL: u,
	}
}

// UnmarshalForm implements custom unmarshler for form values.
func (u *URL) UnmarshalForm(v []byte) error {
	url, err := ParseURL(string(v))
	if err != nil {
		return fmt.Errorf("URL: UnmarshalForm: %w", err)
	}

	*u = *url

	return nil
}

// UnmarshalJSON implements custom unmarshler for JSON.
func (u *URL) UnmarshalJSON(v []byte) error {
	src, err := strconv.Unquote(string(v))
	if err != nil {
		return fmt.Errorf("URL: UnmarshalJSON: %w", err)
	}

	url, err := ParseURL(src)
	if err != nil {
		return fmt.Errorf("URL: UnmarshalJSON: %w", err)
	}

	*u = *url

	return nil
}

func (u URL) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Quote(u.String())), nil
}

func (u URL) GoString() string {
	if u.URL == nil {
		return "domain.URL(" + common.Und + ")"
	}

	return "domain.URL(" + u.URL.String() + ")"
}
