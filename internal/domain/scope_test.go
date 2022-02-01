package domain_test

import (
	"fmt"
	"reflect"
	"testing"

	"source.toby3d.me/website/indieauth/internal/domain"
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

func TestScopes_UnmarshalForm(t *testing.T) {
	t.Parallel()

	input := []byte("profile email")
	results := make(domain.Scopes, 0)

	if err := results.UnmarshalForm(input); err != nil {
		t.Fatalf("%+v", err)
	}

	expResults := domain.Scopes{domain.ScopeProfile, domain.ScopeEmail}
	if !reflect.DeepEqual(results, expResults) {
		t.Errorf("UnmarshalForm(%s) = %s, want %s", input, results, expResults)
	}
}

func TestScopes_UnmarshalJSON(t *testing.T) {
	t.Parallel()

	input := []byte(`"profile email"`)
	results := make(domain.Scopes, 0)

	if err := results.UnmarshalJSON(input); err != nil {
		t.Fatalf("%+v", err)
	}

	expResults := domain.Scopes{domain.ScopeProfile, domain.ScopeEmail}
	if !reflect.DeepEqual(results, expResults) {
		t.Errorf("UnmarshalJSON(%s) = %s, want %s", input, results, expResults)
	}
}

func TestScopes_MarshalJSON(t *testing.T) {
	t.Parallel()

	scopes := domain.Scopes{domain.ScopeEmail, domain.ScopeProfile}

	result, err := scopes.MarshalJSON()
	if err != nil {
		t.Fatalf("%+v", err)
	}

	if string(result) != fmt.Sprintf(`"%s"`, scopes) {
		t.Errorf("MarshalJSON() = %s, want %s", result, fmt.Sprintf(`"%s"`, scopes))
	}
}

func TestScope_String(t *testing.T) {
	t.Parallel()

	//nolint: paralleltest // false positive, in is used
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

func TestScopes_String(t *testing.T) {
	t.Parallel()

	scopes := domain.Scopes{domain.ScopeProfile, domain.ScopeEmail}
	if result := scopes.String(); result != fmt.Sprint(scopes) {
		t.Errorf("String() = %s, want %s", result, scopes)
	}
}

func TestScopes_IsEmpty(t *testing.T) {
	t.Parallel()

	scopes := domain.Scopes{domain.ScopeUndefined}
	if result := scopes.IsEmpty(); !result {
		t.Errorf("IsEmpty() = %t, want %t", result, true)
	}
}

func TestScopes_Has(t *testing.T) {
	t.Parallel()

	scopes := domain.Scopes{domain.ScopeProfile, domain.ScopeEmail}
	if result := scopes.Has(domain.ScopeEmail); !result {
		t.Errorf("Has(%s) = %t, want %t", domain.ScopeEmail, result, true)
	}
}
