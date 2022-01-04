package http_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/fasthttp/router"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	http "github.com/valyala/fasthttp"

	"source.toby3d.me/website/indieauth/internal/common"
	"source.toby3d.me/website/indieauth/internal/domain"
	"source.toby3d.me/website/indieauth/internal/testing/httptest"
	repository "source.toby3d.me/website/indieauth/internal/ticket/repository/http"
)

type TestCase struct {
	name      string
	bodyLinks map[string]string
	metadata  string
}

const testBody string = `
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Secret</title>
    %s
  </head>
  <body>
    <h1>Nothing to see here.</h1>
  </body>
</html>
`

func TestGet(t *testing.T) {
	t.Parallel()

	resource := domain.TestURL(t, "https://alice.example.com/private")
	endpoint := domain.TestURL(t, "https://example.org/token")

	for _, testCase := range []TestCase{{
		name: "link",
		bodyLinks: map[string]string{
			"token_endpoint": endpoint.String(),
		},
		metadata: `{"token_endpoint": ""}`,
	}, {
		name: "metadata",
		bodyLinks: map[string]string{
			"indieauth-metadata": "https://example.com/.well-known/oauth-authorization-server",
		},
		metadata: `{"token_endpoint": "` + endpoint.String() + `"}`,
	}, {
		name: "fallback",
		bodyLinks: map[string]string{
			"token_endpoint":     endpoint.String(),
			"indieauth-metadata": "https://example.com/.well-known/oauth-authorization-server",
		},
		metadata: `{}`,
	}, {
		name: "priority",
		bodyLinks: map[string]string{
			"token_endpoint":     "dont-touch-me",
			"indieauth-metadata": "https://example.com/.well-known/oauth-authorization-server",
		},
		metadata: `{"token_endpoint": "` + endpoint.String() + `"}`,
	}} {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			r := router.New()
			r.GET("/.well-known/oauth-authorization-server", func(ctx *http.RequestCtx) {
				ctx.SuccessString(common.MIMEApplicationJSONCharsetUTF8, testCase.metadata)
			})
			r.GET("/private", func(ctx *http.RequestCtx) {
				bodyLinks := make([]string, 0)
				for k, v := range testCase.bodyLinks {
					bodyLinks = append(bodyLinks, `<link rel="`+k+`" href="`+v+`">`)
				}

				ctx.SuccessString(
					common.MIMETextHTMLCharsetUTF8,
					fmt.Sprintf(testBody, strings.Join(bodyLinks, "\n")),
				)
			})

			client, _, cleanup := httptest.New(t, r.Handler)
			t.Cleanup(cleanup)

			result, err := repository.NewHTTPTicketRepository(client).
				Get(context.Background(), resource)
			require.NoError(t, err)
			assert.Equal(t, endpoint.String(), result.String())
		})
	}
}
