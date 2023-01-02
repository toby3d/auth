package domain

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"inet.af/netaddr"
)

// ClientID is a URL client identifier.
type ClientID struct {
	clientID *url.URL
}

//nolint:gochecknoglobals // slices cannot be constants
var (
	localhostIPv4 = netaddr.MustParseIP("127.0.0.1")
	localhostIPv6 = netaddr.MustParseIP("::1")
)

// ParseClientID parse string as client ID URL identifier.
//
//nolint:funlen,cyclop
func ParseClientID(src string) (*ClientID, error) {
	cid, err := url.Parse(src)
	if err != nil {
		return nil, NewError(
			ErrorCodeInvalidRequest,
			err.Error(),
			"https://indieauth.net/source/#client-identifier",
		)
	}

	if cid.Scheme != "http" && cid.Scheme != "https" {
		return nil, NewError(
			ErrorCodeInvalidRequest,
			"client identifier URL MUST have either an https or http scheme",
			"https://indieauth.net/source/#client-identifier",
		)
	}

	if cid.Path == "" || strings.Contains(cid.Path, "/.") || strings.Contains(cid.Path, "/..") {
		return nil, NewError(
			ErrorCodeInvalidRequest,
			"client identifier URL MUST contain a path component and MUST NOT contain "+
				"single-dot or double-dot path segments",
			"https://indieauth.net/source/#client-identifier",
		)
	}

	if cid.Fragment != "" {
		return nil, NewError(
			ErrorCodeInvalidRequest,
			"client identifier URL MUST NOT contain a fragment component",
			"https://indieauth.net/source/#client-identifier",
		)
	}

	if cid.User != nil {
		return nil, NewError(
			ErrorCodeInvalidRequest,
			"client identifier URL MUST NOT contain a username or password component",
			"https://indieauth.net/source/#client-identifier",
		)
	}

	domain := cid.Hostname()
	if domain == "" {
		return nil, NewError(
			ErrorCodeInvalidRequest,
			"client host name MUST be domain name or a loopback interface",
			"https://indieauth.net/source/#client-identifier",
		)
	}

	ip, err := netaddr.ParseIP(domain)
	if err != nil {
		ipPort, err := netaddr.ParseIPPort(domain)
		if err != nil {
			//nolint:nilerr // ClientID does not contain an IP address, so it is valid
			return &ClientID{clientID: cid}, nil
		}

		ip = ipPort.IP()
	}

	if !ip.IsLoopback() && ip.Compare(localhostIPv4) != 0 && ip.Compare(localhostIPv6) != 0 {
		return nil, NewError(
			ErrorCodeInvalidRequest,
			"client identifier URL MUST NOT be IPv4 or IPv6 addresses except for IPv4 "+
				"127.0.0.1 or IPv6 [::1]",
			"https://indieauth.net/source/#client-identifier",
		)
	}

	return &ClientID{
		clientID: cid,
	}, nil
}

// TestClientID returns valid random generated ClientID for tests.
func TestClientID(tb testing.TB) *ClientID {
	tb.Helper()

	clientID, err := ParseClientID("https://example.com/")
	if err != nil {
		tb.Fatal(err)
	}

	return clientID
}

// UnmarshalForm implements custom unmarshler for form values.
func (cid *ClientID) UnmarshalForm(v []byte) error {
	clientID, err := ParseClientID(string(v))
	if err != nil {
		return fmt.Errorf("ClientID: UnmarshalForm: %w", err)
	}

	*cid = *clientID

	return nil
}

// UnmarshalJSON implements custom unmarshler for JSON.
func (cid *ClientID) UnmarshalJSON(v []byte) error {
	src, err := strconv.Unquote(string(v))
	if err != nil {
		return fmt.Errorf("ClientID: UnmarshalJSON: %w", err)
	}

	clientID, err := ParseClientID(src)
	if err != nil {
		return fmt.Errorf("ClientID: UnmarshalJSON: %w", err)
	}

	*cid = *clientID

	return nil
}

// MarshalForm implements custom marshler for JSON.
func (cid ClientID) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Quote(cid.String())), nil
}

// URL returns url.URL representation of client ID.
func (cid ClientID) URL() *url.URL {
	out, _ := url.Parse(cid.clientID.String())

	return out
}

// String returns string representation of client ID.
func (cid ClientID) String() string {
	return cid.clientID.String()
}
