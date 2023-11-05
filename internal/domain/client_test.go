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
		"prefix":       client.ID.URL().JoinPath("/callback"),
		"redirect_uri": client.RedirectURI[len(client.RedirectURI)-1],
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
