package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	http "github.com/valyala/fasthttp"

	"source.toby3d.me/website/indieauth/internal/domain"
)

func TestClient_ValidateRedirectURI(t *testing.T) {
	t.Parallel()

	client := domain.TestClient(t)

	for _, testCase := range []struct {
		name      string
		input     func() *domain.URL
		expResult bool
	}{{
		name: "client_id prefix",
		input: func() *domain.URL {
			u := &domain.URL{
				URI: http.AcquireURI(),
			}
			client.ID.URI().CopyTo(u.URI)
			u.SetPath("/callback")

			return u
		},
		expResult: true,
	}, {
		name: "registered redirect_uri",
		input: func() *domain.URL {
			return client.RedirectURI[len(client.RedirectURI)-1]
		},
		expResult: true,
	}} {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, testCase.expResult, client.ValidateRedirectURI(testCase.input()))
		})
	}
}
