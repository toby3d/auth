package domain

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"testing"
)

// Me is a URL user identifier.
type Me struct {
	id *url.URL
}

// ParseMe parse string as me URL identifier.
//
//nolint:funlen,cyclop
func ParseMe(raw string) (*Me, error) {
	id, err := url.Parse(raw)
	if err != nil {
		return nil, NewError(
			ErrorCodeInvalidRequest,
			err.Error(),
			"https://indieauth.net/source/#user-profile-url",
			"",
		)
	}

	if id.Scheme != "http" && id.Scheme != "https" {
		return nil, NewError(
			ErrorCodeInvalidRequest,
			"profile URL MUST have either an https or http scheme",
			"https://indieauth.net/source/#user-profile-url",
			"",
		)
	}

	if id.Path == "" {
		id.Path = "/"
	}

	if strings.Contains(id.Path, "/.") || strings.Contains(id.Path, "/..") {
		return nil, NewError(
			ErrorCodeInvalidRequest,
			"profile URL MUST contain a path component (/ is a valid path), MUST NOT contain single-dot "+
				"or double-dot path segments",
			"https://indieauth.net/source/#user-profile-url",
			"",
		)
	}

	if id.Fragment != "" {
		return nil, NewError(
			ErrorCodeInvalidRequest,
			"profile URL MUST NOT contain a fragment component",
			"https://indieauth.net/source/#user-profile-url",
			"",
		)
	}

	if id.User != nil {
		return nil, NewError(
			ErrorCodeInvalidRequest,
			"profile URL MUST NOT contain a username or password component",
			"https://indieauth.net/source/#user-profile-url",
			"",
		)
	}

	if id.Host == "" {
		return nil, NewError(
			ErrorCodeInvalidRequest,
			"profile host name MUST be a domain name",
			"https://indieauth.net/source/#user-profile-url",
			"",
		)
	}

	if _, port, _ := net.SplitHostPort(id.Host); port != "" {
		return nil, NewError(
			ErrorCodeInvalidRequest,
			"profile MUST NOT contain a port",
			"https://indieauth.net/source/#user-profile-url",
			"",
		)
	}

	if net.ParseIP(id.Host) != nil {
		return nil, NewError(
			ErrorCodeInvalidRequest,
			"profile MUST NOT be ipv4 or ipv6 addresses",
			"https://indieauth.net/source/#user-profile-url",
			"",
		)
	}

	return &Me{id: id}, nil
}

// TestMe returns valid random generated me for tests.
func TestMe(tb testing.TB, src string) *Me {
	tb.Helper()

	me, err := ParseMe(src)
	if err != nil {
		tb.Fatal(err)
	}

	return me
}

// UnmarshalForm implements custom unmarshler for form values.
func (m *Me) UnmarshalForm(v []byte) error {
	me, err := ParseMe(string(v))
	if err != nil {
		return fmt.Errorf("Me: UnmarshalForm: %w", err)
	}

	*m = *me

	return nil
}

// UnmarshalJSON implements custom unmarshler for JSON.
func (m *Me) UnmarshalJSON(v []byte) error {
	src, err := strconv.Unquote(string(v))
	if err != nil {
		return fmt.Errorf("Me: UnmarshalJSON: %w", err)
	}

	me, err := ParseMe(src)
	if err != nil {
		return fmt.Errorf("Me: UnmarshalJSON: %w", err)
	}

	*m = *me

	return nil
}

// MarshalJSON implements custom marshler for JSON.
func (m Me) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Quote(m.String())), nil
}

// URL returns copy of parsed me in *url.URL representation.
func (m Me) URL() *url.URL {
	if m.id == nil {
		return nil
	}

	out, _ := url.Parse(m.id.String())

	return out
}

// String returns string representation of me.
func (m Me) String() string {
	if m.id != nil {
		return m.id.String()
	}

	return ""
}

func (m Me) GoString() string {
	return "domain.Me(" + m.String() + ")"
}
