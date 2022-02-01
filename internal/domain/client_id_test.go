package domain_test

import (
	"fmt"
	"testing"

	"source.toby3d.me/website/indieauth/internal/domain"
)

func TestParseClientID(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name     string
		in       string
		expError bool
	}{
		{name: "valid", in: "https://example.com/", expError: false},
		{name: "valid path", in: "https://example.com/username", expError: false},
		{name: "valid query", in: "https://example.com/users?id=100", expError: false},
		{name: "valid port", in: "https://example.com:8443/", expError: false},
		{name: "valid loopback", in: "https://127.0.0.1:8443/", expError: false},
		{name: "missing scheme", in: "example.com", expError: true},
		{name: "invalid scheme", in: "mailto:user@example.com", expError: true},
		{name: "invalid double-dot path", in: "https://example.com/foo/../bar", expError: true},
		{name: "invalid fragment", in: "https://example.com/#me", expError: true},
		{name: "invalid user", in: "https://user:pass@example.com/", expError: true},
		{name: "host is an IP address", in: "https://172.28.92.51/", expError: true},
	} {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := domain.ParseClientID(tc.in)

			switch {
			case err != nil && !tc.expError:
				t.Errorf("ParseClientID(%s) = %+v, want nil", tc.in, err)
			case err == nil && tc.expError:
				t.Errorf("ParseClientID(%s) = %+v, want error", tc.in, err)
			}
		})
	}
}

func TestClientID_UnmarshalForm(t *testing.T) {
	t.Parallel()

	cid := domain.TestClientID(t)
	input := []byte(fmt.Sprint(cid))
	result := new(domain.ClientID)

	if err := result.UnmarshalForm(input); err != nil {
		t.Fatalf("%+v", err)
	}

	if fmt.Sprint(result) != fmt.Sprint(cid) {
		t.Errorf("UnmarshalForm(%s) = %v, want %v", input, result, cid)
	}
}

func TestClientID_UnmarshalJSON(t *testing.T) {
	t.Parallel()

	cid := domain.TestClientID(t)
	input := []byte(fmt.Sprintf(`"%s"`, cid))
	result := new(domain.ClientID)

	if err := result.UnmarshalJSON(input); err != nil {
		t.Fatalf("%+v", err)
	}

	if fmt.Sprint(result) != fmt.Sprint(cid) {
		t.Errorf("UnmarshalJSON(%s) = %v, want %v", input, result, cid)
	}
}

func TestClientID_MarshalJSON(t *testing.T) {
	t.Parallel()

	cid := domain.TestClientID(t)

	result, err := cid.MarshalJSON()
	if err != nil {
		t.Fatalf("%+v", err)
	}

	if string(result) != fmt.Sprintf(`"%s"`, cid) {
		t.Errorf("MarshalJSON() = %s, want %s", result, fmt.Sprintf(`"%s"`, cid))
	}
}

// TODO(toby3d): TestClientID_URI

// TODO(toby3d): TestClientID_URL

func TestClientID_String(t *testing.T) {
	t.Parallel()

	if cid := domain.TestClientID(t); cid.String() != fmt.Sprint(cid) {
		t.Errorf("String() = %s, want %s", cid.String(), fmt.Sprint(cid))
	}
}
