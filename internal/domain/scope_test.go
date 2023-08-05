package domain_test

import (
	"fmt"
	"testing"

	"source.toby3d.me/toby3d/auth/internal/domain"
)

func TestParseScope(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		in  string
		out domain.Scope
	}{
		{in: "create", out: domain.ScopeCreate},
		{in: "delete", out: domain.ScopeDelete},
		{in: "draft", out: domain.ScopeDraft},
		{in: "media", out: domain.ScopeMedia},
		{in: "update", out: domain.ScopeUpdate},
		{in: "block", out: domain.ScopeBlock},
		{in: "channels", out: domain.ScopeChannels},
		{in: "follow", out: domain.ScopeFollow},
		{in: "mute", out: domain.ScopeMute},
		{in: "read", out: domain.ScopeRead},
		{in: "profile", out: domain.ScopeProfile},
		{in: "email", out: domain.ScopeEmail},
	} {
		tc := tc

		t.Run(tc.in, func(t *testing.T) {
			t.Parallel()

			result, err := domain.ParseScope(tc.in)
			if err != nil {
				t.Fatalf("%+v", err)
			}

			if result != tc.out {
				t.Errorf("ParseScope(%s) = %v, want %v", tc.in, result, tc.out)
			}
		})
	}
}

func TestScope_String(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		in  domain.Scope
		out string
	}{
		{in: domain.ScopeCreate, out: "create"},
		{in: domain.ScopeDelete, out: "delete"},
		{in: domain.ScopeDraft, out: "draft"},
		{in: domain.ScopeMedia, out: "media"},
		{in: domain.ScopeUpdate, out: "update"},
		{in: domain.ScopeBlock, out: "block"},
		{in: domain.ScopeChannels, out: "channels"},
		{in: domain.ScopeFollow, out: "follow"},
		{in: domain.ScopeMute, out: "mute"},
		{in: domain.ScopeRead, out: "read"},
		{in: domain.ScopeProfile, out: "profile"},
		{in: domain.ScopeEmail, out: "email"},
	} {
		tc := tc

		t.Run(fmt.Sprint(tc.in), func(t *testing.T) {
			t.Parallel()

			if result := tc.in.String(); result != tc.out {
				t.Errorf("String() = %s, want %s", result, tc.out)
			}
		})
	}
}
