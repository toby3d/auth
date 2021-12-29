package http_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	http "github.com/valyala/fasthttp"

	repository "source.toby3d.me/website/oauth/internal/client/repository/http"
	"source.toby3d.me/website/oauth/internal/common"
	"source.toby3d.me/website/oauth/internal/domain"
	"source.toby3d.me/website/oauth/internal/testing/httptest"
)

const testBody string = `
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%[1]s</title>
    <link rel="redirect_uri" href="%[4]s">
  </head>
  <body>
    <div class="h-app h-x-app">
      <img class="u-logo" src="%[3]s">
      <a class="u-url p-name" href="%[2]s">%[1]s</a>
    </div>
  </body>
</html>
`

func TestGet(t *testing.T) {
	t.Parallel()

	client := domain.TestClient(t)
	httpClient, _, cleanup := httptest.New(t, testHandler(t, client))
	t.Cleanup(cleanup)

	result, err := repository.NewHTTPClientRepository(httpClient).Get(context.TODO(), client.ID)
	require.NoError(t, err)

	assert.Equal(t, client.Name, result.Name)
	assert.Equal(t, client.ID.String(), result.ID.String())

	for i := range client.URL {
		assert.Equal(t, client.URL[i].String(), result.URL[i].String())
	}

	for i := range client.Logo {
		assert.Equal(t, client.Logo[i].String(), result.Logo[i].String())
	}

	for i := range client.RedirectURI {
		assert.Equal(t, client.RedirectURI[i].String(), result.RedirectURI[i].String())
	}
}

func testHandler(tb testing.TB, client *domain.Client) http.RequestHandler {
	tb.Helper()

	return func(ctx *http.RequestCtx) {
		ctx.Response.Header.Set(http.HeaderLink, `<`+client.RedirectURI[0].String()+`>; rel="redirect_uri"`)
		ctx.SuccessString(common.MIMETextHTMLCharsetUTF8, fmt.Sprintf(
			testBody, client.Name[0], client.URL[0].String(), client.Logo[0].String(),
			client.RedirectURI[1].String(),
		))
	}
}
