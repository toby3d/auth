package domain_test

import (
	"net/url"
	"testing"

	"source.toby3d.me/toby3d/auth/internal/domain"
)

func TestClient_ValidateRedirectURI(t *testing.T) {
	t.Parallel()

	client := domain.TestClient(t)

	for name, in := range map[string]*url.URL{
		"client_id prefix":        client.ID.URL().JoinPath("/callback"),
		"registered redirect_uri": client.RedirectURI[len(client.RedirectURI)-1],
	} {
		name, in := name, in

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if out := client.ValidateRedirectURI(in); !out {
				t.Errorf("ValidateRedirectURI(%v) = %t, want %t", in, out, true)
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
