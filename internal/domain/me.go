package domain

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"testing"

	http "github.com/valyala/fasthttp"
)

// Me is a URL user identifier.
type Me struct {
	me *http.URI
}

// ParseMe parse string as me URL identifier.
//nolint: funlen, cyclop
func ParseMe(raw string) (*Me, error) {
	me := http.AcquireURI()
	if err := me.Parse(nil, []byte(raw)); err != nil {
		return nil, NewError(
			ErrorCodeInvalidRequest,
			err.Error(),
			"https://indieauth.net/source/#user-profile-url",
			"",
		)
	}

	scheme := string(me.Scheme())
	if scheme != "http" && scheme != "https" {
		return nil, NewError(
			ErrorCodeInvalidRequest,
			"profile URL MUST have either an https or http scheme",
			"https://indieauth.net/source/#user-profile-url",
			"",
		)
	}

	path := string(me.PathOriginal())
	if path == "" || strings.Contains(path, "/.") || strings.Contains(path, "/..") {
		return nil, NewError(
			ErrorCodeInvalidRequest,
			"profile URL MUST contain a path component (/ is a valid path), MUST NOT contain single-dot "+
				"or double-dot path segments",
			"https://indieauth.net/source/#user-profile-url",
			"",
		)
	}

	if me.Hash() != nil {
		return nil, NewError(
			ErrorCodeInvalidRequest,
			"profile URL MUST NOT contain a fragment component",
			"https://indieauth.net/source/#user-profile-url",
			"",
		)
	}

	if me.Username() != nil || me.Password() != nil {
		return nil, NewError(
			ErrorCodeInvalidRequest,
			"profile URL MUST NOT contain a username or password component",
			"https://indieauth.net/source/#user-profile-url",
			"",
		)
	}

	domain := string(me.Host())
	if domain == "" {
		return nil, NewError(
			ErrorCodeInvalidRequest,
			"profile host name MUST be a domain name",
			"https://indieauth.net/source/#user-profile-url",
			"",
		)
	}

	if _, port, _ := net.SplitHostPort(domain); port != "" {
		return nil, NewError(
			ErrorCodeInvalidRequest,
			"profile MUST NOT contain a port",
			"https://indieauth.net/source/#user-profile-url",
			"",
		)
	}

	if net.ParseIP(domain) != nil {
		return nil, NewError(
			ErrorCodeInvalidRequest,
			"profile MUST NOT be ipv4 or ipv6 addresses",
			"https://indieauth.net/source/#user-profile-url",
			"",
		)
	}

	return &Me{me: me}, nil
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
		return fmt.Errorf("UnmarshalForm: %w", err)
	}

	*m = *me

	return nil
}

// UnmarshalJSON implements custom unmarshler for JSON.
func (m *Me) UnmarshalJSON(v []byte) error {
	src, err := strconv.Unquote(string(v))
	if err != nil {
		return fmt.Errorf("UnmarshalJSON: %w", err)
	}

	me, err := ParseMe(src)
	if err != nil {
		return fmt.Errorf("UnmarshalJSON: %w", err)
	}

	*m = *me

	return nil
}

// MarshalJSON implements custom marshler for JSON.
func (m Me) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Quote(m.String())), nil
}

// URI returns copy of parsed me in *fasthttp.URI representation.
// This copy MUST be released via fasthttp.ReleaseURI.
func (m Me) URI() *http.URI {
	if m.me == nil {
		return nil
	}

	u := http.AcquireURI()
	m.me.CopyTo(u)

	return u
}

// URL returns copy of parsed me in *url.URL representation.
func (m Me) URL() *url.URL {
	if m.me == nil {
		return nil
	}

	return &url.URL{
		ForceQuery:  false,
		Fragment:    string(m.me.Hash()),
		Host:        string(m.me.Host()),
		Opaque:      "",
		Path:        string(m.me.Path()),
		RawFragment: "",
		RawPath:     string(m.me.PathOriginal()),
		RawQuery:    string(m.me.QueryString()),
		Scheme:      string(m.me.Scheme()),
		User:        nil,
	}
}

// String returns string representation of me.
func (m Me) String() string {
	if m.me == nil {
		return ""
	}

	return m.me.String()
}
