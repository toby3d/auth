package domain_test

import (
	"fmt"
	"testing"

	"source.toby3d.me/website/indieauth/internal/domain"
)

func TestParseURL(t *testing.T) {
	t.Parallel()

	input := "https://user:pass@example.com:8443/users?id=100#me"
	if _, err := domain.ParseURL(input); err != nil {
		t.Errorf("ParseURL(%s) = %+v, want nil", input, err)
	}
}

func TestURL_UnmarshalForm(t *testing.T) {
	t.Parallel()

	u := domain.TestURL(t, "https://user:pass@example.com:8443/users?id=100#me")
	input := []byte(fmt.Sprint(u))
	result := new(domain.URL)

	if err := result.UnmarshalForm(input); err != nil {
		t.Fatalf("%+v", err)
	}

	if fmt.Sprint(result) != fmt.Sprint(u) {
		t.Errorf("UnmarshalForm(%s) = %v, want %v", input, result, u)
	}
}

func TestURL_UnmarshalJSON(t *testing.T) {
	t.Parallel()

	u := domain.TestURL(t, "https://user:pass@example.com:8443/users?id=100#me")
	input := []byte(fmt.Sprintf(`"%s"`, u))
	result := new(domain.URL)

	if err := result.UnmarshalJSON(input); err != nil {
		t.Fatalf("%+v", err)
	}

	if fmt.Sprint(result) != fmt.Sprint(u) {
		t.Errorf("UnmarshalJSON(%s) = %v, want %v", input, result, u)
	}
}

// TODO(toby3d): TestURL_URL
