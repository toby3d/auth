package domain_test

import (
	"fmt"
	"testing"

	"source.toby3d.me/website/indieauth/internal/domain"
)

//nolint: funlen
func TestParseMe(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name     string
		in       string
		expError bool
	}{{
		name:     "valid",
		in:       "https://example.com/",
		expError: false,
	}, {
		name:     "valid path",
		in:       "https://example.com/username",
		expError: false,
	}, {
		name:     "valid query",
		in:       "https://example.com/users?id=100",
		expError: false,
	}, {
		name:     "missing scheme",
		in:       "example.com",
		expError: true,
	}, {
		name:     "invalid scheme",
		in:       "mailto:user@example.com",
		expError: true,
	}, {
		name:     "contains double-dot path",
		in:       "https://example.com/foo/../bar",
		expError: true,
	}, {
		name:     "contains fragment",
		in:       "https://example.com/#me",
		expError: true,
	}, {
		name:     "contains user",
		in:       "https://user:pass@example.com/",
		expError: true,
	}, {
		name:     "contains port",
		in:       "https://example.com:8443/",
		expError: true,
	}, {
		name:     "host is an IP address",
		in:       "https://172.28.92.51/",
		expError: true,
	}} {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := domain.ParseMe(tc.in)

			switch {
			case err != nil && !tc.expError:
				t.Errorf("ParseMe(%s) = %+v, want nil", tc.in, err)
			case err == nil && tc.expError:
				t.Errorf("ParseMe(%s) = %+v, want error", tc.in, err)
			}
		})
	}
}

func TestMe_UnmarshalForm(t *testing.T) {
	t.Parallel()

	me := domain.TestMe(t, "https://user.example.com/")
	input := []byte(fmt.Sprint(me))
	result := new(domain.Me)

	if err := result.UnmarshalForm(input); err != nil {
		t.Fatalf("%+v", err)
	}

	if fmt.Sprint(result) != fmt.Sprint(me) {
		t.Errorf("UnmarshalForm(%s) = %v, want %v", input, result, me)
	}
}

func TestMe_UnmarshalJSON(t *testing.T) {
	t.Parallel()

	me := domain.TestMe(t, "https://user.example.com/")
	input := []byte(fmt.Sprintf(`"%s"`, me))
	result := new(domain.Me)

	if err := result.UnmarshalJSON(input); err != nil {
		t.Fatalf("%+v", err)
	}

	if fmt.Sprint(result) != fmt.Sprint(me) {
		t.Errorf("UnmarshalJSON(%s) = %v, want %v", input, result, me)
	}
}

func TestMe_MarshalJSON(t *testing.T) {
	t.Parallel()

	me := domain.TestMe(t, "https://user.example.com/")

	result, err := me.MarshalJSON()
	if err != nil {
		t.Fatalf("%+v", err)
	}

	if string(result) != fmt.Sprintf(`"%s"`, me) {
		t.Errorf("MarshalJSON() = %s, want %s", result, fmt.Sprintf(`"%s"`, me))
	}
}

// TODO(toby3d): TestMe_URI

// TODO(toby3d): TestMe_URL

func TestMe_String(t *testing.T) {
	t.Parallel()

	me := domain.TestMe(t, "https://user.example.com/")
	if result := me.String(); result != fmt.Sprint(me) {
		t.Errorf("Strig() = %s, want %s", result, fmt.Sprint(me))
	}
}
