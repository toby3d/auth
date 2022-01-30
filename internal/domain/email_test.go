package domain_test

import (
	"fmt"
	"testing"

	"source.toby3d.me/website/indieauth/internal/domain"
)

func TestParseEmail(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		in   string
		out  string
	}{{
		name: "simple",
		in:   "user@example.com",
		out:  "user@example.com",
	}, {
		name: "subAddress",
		in:   "user+suffix@example.com",
		out:  "user+suffix@example.com",
	}, {
		name: "mailto prefix",
		in:   "mailto:user@example.com",
		out:  "user@example.com",
	}} {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result, err := domain.ParseEmail(tc.in)
			if err != nil {
				t.Fatalf("%+v", err)
			}

			if fmt.Sprint(result) != tc.out {
				t.Errorf("ParseEmail(%s) = %s, want %s", tc.in, result, tc.out)
			}
		})
	}
}

func TestEmail_String(t *testing.T) {
	t.Parallel()

	email := domain.TestEmail(t)
	if result := email.String(); result != fmt.Sprint(email) {
		t.Errorf("String() = %v, want %v", result, email)
	}
}
