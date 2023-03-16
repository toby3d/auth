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
	me *url.URL
}

// ParseMe parse string as me URL identifier.
//
//nolint:funlen,cyclop
func ParseMe(raw string) (*Me, error) {
	me, err := url.Parse(raw)
	if err != nil {
		return nil, NewError(
			ErrorCodeInvalidRequest,
			err.Error(),
			"https://indieauth.net/source/#user-profile-url",
			"",
		)
	}

	if me.Scheme != "http" && me.Scheme != "https" {
		return nil, NewError(
			ErrorCodeInvalidRequest,
			"profile URL MUST have either an https or http scheme, got '"+me.Scheme+"'",
			"https://indieauth.net/source/#user-profile-url",
			"",
		)
	}

	if me.Path == "" {
		me.Path = "/"
	}

	if strings.Contains(me.Path, "/.") || strings.Contains(me.Path, "/..") {
		return nil, NewError(
			ErrorCodeInvalidRequest,
			"profile URL MUST contain a path component (/ is a valid path), MUST NOT contain single-dot "+
				"or double-dot path segments, got '"+me.Path+"'",
			"https://indieauth.net/source/#user-profile-url",
			"",
		)
	}

	if me.Fragment != "" {
		return nil, NewError(
			ErrorCodeInvalidRequest,
			"profile URL MUST NOT contain a fragment component, got '"+me.Fragment+"'",
			"https://indieauth.net/source/#user-profile-url",
			"",
		)
	}

	if me.User != nil {
		return nil, NewError(
			ErrorCodeInvalidRequest,
			"profile URL MUST NOT contain a username or password component, got '"+me.User.String()+"'",
			"https://indieauth.net/source/#user-profile-url",
			"",
		)
	}

	if me.Host == "" {
		return nil, NewError(
			ErrorCodeInvalidRequest,
			"profile host name MUST be a domain name, got '"+me.Host+"'",
			"https://indieauth.net/source/#user-profile-url",
			"",
		)
	}

	if _, port, _ := net.SplitHostPort(me.Host); port != "" {
		return nil, NewError(
			ErrorCodeInvalidRequest,
			"profile MUST NOT contain a port, got '"+port+"'",
			"https://indieauth.net/source/#user-profile-url",
			"",
		)
	}

	if out := net.ParseIP(me.Host); out != nil {
		return nil, NewError(
			ErrorCodeInvalidRequest,
			"profile MUST NOT be ipv4 or ipv6 addresses, got '"+out.String()+"'",
			"https://indieauth.net/source/#user-profile-url",
			"",
		)
	}

	return &Me{me: me}, nil
}

// TestMe returns valid random generated me for tests.
func TestMe(tb testing.TB, src string) *Me {
	tb.Helper()

	u, err := url.Parse(src)
	if err != nil {
		tb.Fatal(err)
	}

	return &Me{me: u}
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
	if m.me == nil {
		return nil
	}

	out, _ := url.Parse(m.me.String())

	return out
}

// String returns string representation of me.
func (m Me) String() string {
	if m.me != nil {
		return m.me.String()
	}

	return ""
}

func (m Me) GoString() string {
	return "domain.Me(" + m.String() + ")"
}
