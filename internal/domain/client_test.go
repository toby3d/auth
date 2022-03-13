package domain_test

import (
	"fmt"
	"testing"

	"source.toby3d.me/toby3d/auth/internal/domain"
)

func TestClient_ValidateRedirectURI(t *testing.T) {
	t.Parallel()

	client := domain.TestClient(t)

	for _, tc := range []struct {
		name string
		in   *domain.URL
	}{
		{name: "client_id prefix", in: domain.TestURL(t, fmt.Sprint(client.ID, "/callback"))},
		{name: "registered redirect_uri", in: client.RedirectURI[len(client.RedirectURI)-1]},
	} {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if result := client.ValidateRedirectURI(tc.in); !result {
				t.Errorf("ValidateRedirectURI(%v) = %t, want %t", tc.in, result, true)
			}
		})
	}
}

func TestClient_GetName(t *testing.T) {
	t.Parallel()

	client := domain.TestClient(t)
	if result := client.GetName(); result != client.Name[0] {
		t.Errorf("GetName() = %v, want %v", result, client.Name[0])
	}
}

func TestClient_GetURL(t *testing.T) {
	t.Parallel()

	client := domain.TestClient(t)
	if result := client.GetURL(); result != client.URL[0] {
		t.Errorf("GetURL() = %v, want %v", result, client.URL[0])
	}
}

func TestClient_GetLogo(t *testing.T) {
	t.Parallel()

	client := domain.TestClient(t)
	if result := client.GetLogo(); result != client.Logo[0] {
		t.Errorf("GetLogo() = %v, want %v", result, client.Logo[0])
	}
}
