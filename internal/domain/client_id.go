package domain

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	http "github.com/valyala/fasthttp"
	"golang.org/x/xerrors"
	"inet.af/netaddr"
)

// ClientID is a URL client identifier.
type ClientID struct {
	clientID *http.URI
	valid    bool
}

//nolint: gochecknoglobals
var (
	localhostIPv4 = netaddr.MustParseIP("127.0.0.1")
	localhostIPv6 = netaddr.MustParseIP("::1")
)

func NewClientID(raw string) (*ClientID, error) {
	clientID := http.AcquireURI()
	if err := clientID.Parse(nil, []byte(raw)); err != nil {
		return nil, Error{
			Code:        "invalid_request",
			Description: err.Error(),
			URI:         "https://indieauth.net/source/#client-identifier",
			Frame:       xerrors.Caller(1),
		}
	}

	scheme := string(clientID.Scheme())
	if scheme != "http" && scheme != "https" {
		return nil, Error{
			Code:        "invalid_request",
			Description: "client identifier URL MUST have either an https or http scheme",
			URI:         "https://indieauth.net/source/#client-identifier",
			Frame:       xerrors.Caller(1),
		}
	}

	path := string(clientID.PathOriginal())
	if path == "" || strings.Contains(path, "/.") || strings.Contains(path, "/..") {
		return nil, Error{
			Code: "invalid_request",
			Description: "client identifier URL MUST contain a path component and MUST NOT contain " +
				"single-dot or double-dot path segments",
			URI:   "https://indieauth.net/source/#client-identifier",
			Frame: xerrors.Caller(1),
		}
	}

	if clientID.Hash() != nil {
		return nil, Error{
			Code:        "invalid_request",
			Description: "client identifier URL MUST NOT contain a fragment component",
			URI:         "https://indieauth.net/source/#client-identifier",
			Frame:       xerrors.Caller(1),
		}
	}

	if clientID.Username() != nil || clientID.Password() != nil {
		return nil, Error{
			Code:        "invalid_request",
			Description: "client identifier URL MUST NOT contain a username or password component",
			URI:         "https://indieauth.net/source/#client-identifier",
			Frame:       xerrors.Caller(1),
		}
	}

	domain := string(clientID.Host())
	if domain == "" {
		return nil, Error{
			Code:        "invalid_request",
			Description: "client host name MUST be domain name or a loopback interface",
			URI:         "https://indieauth.net/source/#client-identifier",
			Frame:       xerrors.Caller(1),
		}
	}

	ip, err := netaddr.ParseIP(domain)
	if err != nil {
		ipPort, err := netaddr.ParseIPPort(domain)
		if err != nil {
			return &ClientID{clientID: clientID}, nil
		}

		ip = ipPort.IP()
	}

	if !ip.IsLoopback() && ip.Compare(localhostIPv4) != 0 && ip.Compare(localhostIPv6) != 0 {
		return nil, Error{
			Code: "invalid_request",
			Description: "client identifier URL MUST NOT be IPv4 or IPv6 addresses except for IPv4 " +
				"127.0.0.1 or IPv6 [::1]",
			URI:   "https://indieauth.net/source/#client-identifier",
			Frame: xerrors.Caller(1),
		}
	}

	return &ClientID{clientID: clientID}, nil
}

// TestClientID returns a valid random generated ClientID for tests.
func TestClientID(tb testing.TB) *ClientID {
	tb.Helper()

	clientID, err := NewClientID("https://app.example.com/")
	require.NoError(tb, err)

	return clientID
}

// UnmarshalForm implements a custom form.Unmarshaler.
func (cid *ClientID) UnmarshalForm(v []byte) error {
	clientID, err := NewClientID(string(v))
	if err != nil {
		return fmt.Errorf("UnmarshalForm: %w", err)
	}

	*cid = *clientID

	return nil
}

func (cid *ClientID) UnmarshalJSON(v []byte) error {
	src, err := strconv.Unquote(string(v))
	if err != nil {
		return err
	}

	clientID, err := NewClientID(src)
	if err != nil {
		return fmt.Errorf("UnmarshalJSON: %w", err)
	}

	*cid = *clientID

	return nil
}

// URI returns copy of parsed *fasthttp.URI.
// This copy MUST be released via fasthttp.ReleaseURI.
func (cid *ClientID) URI() *http.URI {
	u := http.AcquireURI()
	cid.clientID.CopyTo(u)

	return u
}

func (cid *ClientID) URL() *url.URL {
	return &url.URL{
		Scheme:   string(cid.clientID.Scheme()),
		Host:     string(cid.clientID.Host()),
		Path:     string(cid.clientID.Path()),
		RawPath:  string(cid.clientID.PathOriginal()),
		RawQuery: string(cid.clientID.QueryString()),
		Fragment: string(cid.clientID.Hash()),
	}
}

// String returns string representation of client ID.
func (cid *ClientID) String() string {
	return cid.clientID.String()
}
