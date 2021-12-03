package domain

import (
	"fmt"
	"net"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	http "github.com/valyala/fasthttp"
)

// Me is a user URL identifier.
type Me struct {
	uri     *http.URI
	isValid bool
}

// UnmarshalForm implements a custom form.Unmarshaler.
func (me *Me) UnmarshalForm(v []byte) error {
	if err := me.Parse(v); err != nil {
		return fmt.Errorf("cannot unmarshal form: %w", err)
	}

	return nil
}

// Parse parse and validate me identifier.
func (me *Me) Parse(v []byte) error {
	if me.uri != nil {
		http.ReleaseURI(me.uri)
	}

	me.uri = http.AcquireURI()
	if err := me.uri.Parse(nil, v); err != nil {
		return fmt.Errorf("cannot parse me: %w", err)
	}

	// NOTE(toby3d): MUST have either an https or http scheme
	scheme := string(me.uri.Scheme())
	if scheme != "http" && scheme != "https" {
		return nil
	}

	// NOTE(toby3d): MUST contain a path component (/ is a valid path)
	// NOTE(toby3d): MUST NOT contain single-dot or double-dot path segments
	path := string(me.uri.PathOriginal())
	if path == "" || strings.Contains(path, "/.") || strings.Contains(path, "/..") {
		return nil
	}

	// NOTE(toby3d): MUST NOT contain a fragment component
	if me.uri.Hash() != nil {
		return nil
	}

	// NOTE(toby3d): MUST NOT contain a username or password component
	if me.uri.Username() != nil || me.uri.Password() != nil {
		return nil
	}

	// NOTE(toby3d): host names MUST be domain names
	host := string(me.uri.Host())
	if host == "" {
		return nil
	}

	// NOTE(toby3d): MUST NOT contain a port
	if _, _, err := net.SplitHostPort(host); err == nil {
		return nil
	}

	// NOTE(toby3d): MUST NOT be ipv4 or ipv6 addresses
	if net.ParseIP(host) != nil {
		return nil
	}

	me.isValid = true

	return nil
}

// String returns string representation of Me.
func (me *Me) String() string {
	if me.uri == nil {
		return ""
	}

	return me.uri.String()
}

// URI returns copy of parsed *fasthttp.URI.
// This copy MUST be released via fasthttp.ReleaseURI.
func (me *Me) URI() *http.URI {
	u := http.AcquireURI()
	me.uri.CopyTo(u)

	return u
}

// IsValid returns true if Me is a valid identifier.
func (me *Me) IsValid() bool {
	return me.isValid
}

// TestMe returns a valid testing Me.
func TestMe(tb testing.TB) *Me {
	tb.Helper()

	me := new(Me)
	require.NoError(tb, me.Parse([]byte("https://user.example.net/")))

	return me
}
