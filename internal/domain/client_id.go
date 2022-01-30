package domain

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"testing"

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

// ParseClientID parse string as client ID URL identifier.
//nolint: funlen
func ParseClientID(src string) (*ClientID, error) {
	cid := http.AcquireURI()
	if err := cid.Parse(nil, []byte(src)); err != nil {
		return nil, Error{
			Code:        ErrorCodeInvalidRequest,
			Description: err.Error(),
			URI:         "https://indieauth.net/source/#client-identifier",
			State:       "",
			frame:       xerrors.Caller(1),
		}
	}

	scheme := string(cid.Scheme())
	if scheme != "http" && scheme != "https" {
		return nil, Error{
			Code:        ErrorCodeInvalidRequest,
			Description: "client identifier URL MUST have either an https or http scheme",
			URI:         "https://indieauth.net/source/#client-identifier",
			State:       "",
			frame:       xerrors.Caller(1),
		}
	}

	path := string(cid.PathOriginal())
	if path == "" || strings.Contains(path, "/.") || strings.Contains(path, "/..") {
		return nil, Error{
			Code: ErrorCodeInvalidRequest,
			Description: "client identifier URL MUST contain a path component and MUST NOT contain " +
				"single-dot or double-dot path segments",
			URI:   "https://indieauth.net/source/#client-identifier",
			State: "",
			frame: xerrors.Caller(1),
		}
	}

	if cid.Hash() != nil {
		return nil, Error{
			Code:        ErrorCodeInvalidRequest,
			Description: "client identifier URL MUST NOT contain a fragment component",
			URI:         "https://indieauth.net/source/#client-identifier",
			State:       "",
			frame:       xerrors.Caller(1),
		}
	}

	if cid.Username() != nil || cid.Password() != nil {
		return nil, Error{
			Code:        ErrorCodeInvalidRequest,
			Description: "client identifier URL MUST NOT contain a username or password component",
			URI:         "https://indieauth.net/source/#client-identifier",
			State:       "",
			frame:       xerrors.Caller(1),
		}
	}

	domain := string(cid.Host())
	if domain == "" {
		return nil, Error{
			Code:        ErrorCodeInvalidRequest,
			Description: "client host name MUST be domain name or a loopback interface",
			URI:         "https://indieauth.net/source/#client-identifier",
			State:       "",
			frame:       xerrors.Caller(1),
		}
	}

	ip, err := netaddr.ParseIP(domain)
	if err != nil {
		ipPort, err := netaddr.ParseIPPort(domain)
		if err != nil {
			return &ClientID{
				clientID: cid,
			}, nil
		}

		ip = ipPort.IP()
	}

	if !ip.IsLoopback() && ip.Compare(localhostIPv4) != 0 && ip.Compare(localhostIPv6) != 0 {
		return nil, Error{
			Code: ErrorCodeInvalidRequest,
			Description: "client identifier URL MUST NOT be IPv4 or IPv6 addresses except for IPv4 " +
				"127.0.0.1 or IPv6 [::1]",
			URI:   "https://indieauth.net/source/#client-identifier",
			State: "",
			frame: xerrors.Caller(1),
		}
	}

	return &ClientID{
		clientID: cid,
	}, nil
}

// TestClientID returns valid random generated ClientID for tests.
func TestClientID(tb testing.TB) *ClientID {
	tb.Helper()

	clientID, err := ParseClientID("https://app.example.com/")
	if err != nil {
		tb.Fatalf("%+v", err)
	}

	return clientID
}

// UnmarshalForm implements custom unmarshler for form values.
func (cid *ClientID) UnmarshalForm(v []byte) error {
	clientID, err := ParseClientID(string(v))
	if err != nil {
		return fmt.Errorf("UnmarshalForm: %w", err)
	}

	*cid = *clientID

	return nil
}

// UnmarshalJSON implements custom unmarshler for JSON.
func (cid *ClientID) UnmarshalJSON(v []byte) error {
	src, err := strconv.Unquote(string(v))
	if err != nil {
		return fmt.Errorf("UnmarshalJSON: %w", err)
	}

	clientID, err := ParseClientID(src)
	if err != nil {
		return fmt.Errorf("UnmarshalJSON: %w", err)
	}

	*cid = *clientID

	return nil
}

// MarshalForm implements custom marshler for JSON.
func (cid ClientID) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Quote(cid.String())), nil
}

// URI returns copy of parsed *fasthttp.URI.
// This copy MUST be released via fasthttp.ReleaseURI.
func (cid ClientID) URI() *http.URI {
	u := http.AcquireURI()
	cid.clientID.CopyTo(u)

	return u
}

// URL returns url.URL representation of client ID.
func (cid ClientID) URL() *url.URL {
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
func (cid ClientID) String() string {
	return cid.clientID.String()
}
