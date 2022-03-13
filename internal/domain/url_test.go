package domain_test

import (
	"fmt"
	"testing"

	"source.toby3d.me/toby3d/auth/internal/domain"
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

	url := domain.TestURL(t, "https://user:pass@example.com:8443/users?id=100#me")
	input := []byte(fmt.Sprint(url))
	result := new(domain.URL)

	if err := result.UnmarshalForm(input); err != nil {
		t.Fatalf("%+v", err)
	}

	if fmt.Sprint(result) != fmt.Sprint(url) {
		t.Errorf("UnmarshalForm(%s) = %v, want %v", input, result, url)
	}
}

func TestURL_UnmarshalJSON(t *testing.T) {
	t.Parallel()

	url := domain.TestURL(t, "https://user:pass@example.com:8443/users?id=100#me")
	input := []byte(fmt.Sprintf(`"%s"`, url))
	result := new(domain.URL)

	if err := result.UnmarshalJSON(input); err != nil {
		t.Fatalf("%+v", err)
	}

	if fmt.Sprint(result) != fmt.Sprint(url) {
		t.Errorf("UnmarshalJSON(%s) = %v, want %v", input, result, url)
	}
}

// TODO(toby3d): TestURL_URL
