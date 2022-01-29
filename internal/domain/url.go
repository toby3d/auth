package domain

import (
	"fmt"
	"net/url"
	"strconv"
	"testing"

	http "github.com/valyala/fasthttp"
)

// URL describe any valid HTTP URL.
type URL struct {
	*http.URI
}

// ParseURL parse strings as URL.
func ParseURL(src string) (*URL, error) {
	u := http.AcquireURI()
	if err := u.Parse(nil, []byte(src)); err != nil {
		return nil, fmt.Errorf("cannot parse URL: %w", err)
	}

	return &URL{URI: u}, nil
}

// TestURL returns URL of provided input for tests.
func TestURL(tb testing.TB, src string) *URL {
	tb.Helper()

	u := http.AcquireURI()
	u.Update(src)

	return &URL{
		URI: u,
	}
}

// UnmarshalForm implements custom unmarshler for form values.
func (u *URL) UnmarshalForm(v []byte) error {
	url, err := ParseURL(string(v))
	if err != nil {
		return fmt.Errorf("UnmarshalForm: %w", err)
	}

	*u = *url

	return nil
}

// UnmarshalJSON implements custom unmarshler for JSON.
func (u *URL) UnmarshalJSON(v []byte) error {
	src, err := strconv.Unquote(string(v))
	if err != nil {
		return fmt.Errorf("UnmarshalJSON: %w", err)
	}

	url, err := ParseURL(src)
	if err != nil {
		return fmt.Errorf("UnmarshalJSON: %w", err)
	}

	*u = *url

	return nil
}

// URL returns url.URL representation of URL.
func (u URL) URL() *url.URL {
	if u.URI == nil {
		return nil
	}

	result, err := url.ParseRequestURI(u.URI.String())
	if err != nil {
		return nil
	}

	return result
}
