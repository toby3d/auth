package domain_test

import (
	"fmt"
	"reflect"
	"testing"

	"source.toby3d.me/toby3d/auth/internal/domain"
)

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

func TestScopes_String(t *testing.T) {
	t.Parallel()

	scopes := domain.Scopes{domain.ScopeProfile, domain.ScopeEmail}
	if result := scopes.String(); result != fmt.Sprint(scopes) {
		t.Errorf("String() = %s, want %s", result, scopes)
	}
}

func TestScopes_IsEmpty(t *testing.T) {
	t.Parallel()

	scopes := domain.Scopes{domain.ScopeUnd}
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
