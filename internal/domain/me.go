package domain

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	http "github.com/valyala/fasthttp"
	"golang.org/x/xerrors"
)

// Me is a URL user identifier.
type Me struct {
	me *http.URI
}

//nolint: funlen
func NewMe(raw string) (*Me, error) {
	me := http.AcquireURI()
	if err := me.Parse(nil, []byte(raw)); err != nil {
		return nil, Error{
			Code:        "invalid_request",
			Description: err.Error(),
			URI:         "https://indieauth.net/source/#user-profile-url",
			Frame:       xerrors.Caller(1),
		}
	}

	scheme := string(me.Scheme())
	if scheme != "http" && scheme != "https" {
		return nil, Error{
			Code:        "invalid_request",
			Description: "profile URL MUST have either an https or http scheme",
			URI:         "https://indieauth.net/source/#user-profile-url",
			Frame:       xerrors.Caller(1),
		}
	}

	path := string(me.PathOriginal())
	if path == "" || strings.Contains(path, "/.") || strings.Contains(path, "/..") {
		return nil, Error{
			Code: "invalid_request",
			Description: "profile URL MUST contain a path component (/ is a valid path), MUST NOT " +
				"contain single-dot or double-dot path segments",
			URI:   "https://indieauth.net/source/#user-profile-url",
			Frame: xerrors.Caller(1),
		}
	}

	if me.Hash() != nil {
		return nil, Error{
			Code:        "invalid_request",
			Description: "profile URL MUST NOT contain a fragment component",
			URI:         "https://indieauth.net/source/#user-profile-url",
			Frame:       xerrors.Caller(1),
		}
	}

	if me.Username() != nil || me.Password() != nil {
		return nil, Error{
			Code:        "invalid_request",
			Description: "profile URL MUST NOT contain a username or password component",
			URI:         "https://indieauth.net/source/#user-profile-url",
			Frame:       xerrors.Caller(1),
		}
	}

	domain := string(me.Host())
	if domain == "" {
		return nil, Error{
			Code:        "invalid_request",
			Description: "profile host name MUST be a domain name",
			URI:         "https://indieauth.net/source/#user-profile-url",
			Frame:       xerrors.Caller(1),
		}
	}

	if _, port, _ := net.SplitHostPort(domain); port != "" {
		return nil, Error{
			Code:        "invalid_request",
			Description: "profile MUST NOT contain a port",
			URI:         "https://indieauth.net/source/#user-profile-url",
			Frame:       xerrors.Caller(1),
		}
	}

	if net.ParseIP(domain) != nil {
		return nil, Error{
			Code:        "invalid_request",
			Description: "profile MUST NOT be ipv4 or ipv6 addresses",
			URI:         "https://indieauth.net/source/#user-profile-url",
			Frame:       xerrors.Caller(1),
		}
	}

	return &Me{me: me}, nil
}

// TestMe returns a valid random generated Me for tests.
func TestMe(tb testing.TB) *Me {
	tb.Helper()

	me, err := NewMe("https://user.example.net/")
	require.NoError(tb, err)

	return me
}

// UnmarshalForm parses the value of the form key into the Me domain.
func (m *Me) UnmarshalForm(v []byte) error {
	me, err := NewMe(string(v))
	if err != nil {
		return fmt.Errorf("UnmarshalForm: %w", err)
	}

	*m = *me

	return nil
}

func (m *Me) UnmarshalJSON(v []byte) error {
	src, err := strconv.Unquote(string(v))
	if err != nil {
		return err
	}

	me, err := NewMe(src)
	if err != nil {
		return fmt.Errorf("UnmarshalForm: %w", err)
	}

	*m = *me

	return nil
}

func (m Me) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Quote(m.String())), nil
}

// URI returns copy of parsed Me in *fasthttp.URI representation.
// This copy MUST be released via fasthttp.ReleaseURI.
func (m Me) URI() *http.URI {
	if m.me == nil {
		return nil
	}

	u := http.AcquireURI()
	m.me.CopyTo(u)

	return u
}

// URL returns copy of parsed Me in *url.URL representation.
func (m Me) URL() *url.URL {
	if m.me == nil {
		return nil
	}

	return &url.URL{
		Scheme:   string(m.me.Scheme()),
		Host:     string(m.me.Host()),
		Path:     string(m.me.Path()),
		RawPath:  string(m.me.PathOriginal()),
		RawQuery: string(m.me.QueryString()),
		Fragment: string(m.me.Hash()),
	}
}

// String returns string representation of Me.
func (m Me) String() string {
	if m.me == nil {
		return ""
	}

	return m.me.String()
}
