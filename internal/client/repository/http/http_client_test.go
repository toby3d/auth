package http_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	http "github.com/valyala/fasthttp"

	repository "source.toby3d.me/website/oauth/internal/client/repository/http"
	"source.toby3d.me/website/oauth/internal/common"
	"source.toby3d.me/website/oauth/internal/domain"
	"source.toby3d.me/website/oauth/internal/util"
)

const testBody string = `
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Example App</title>
    <link rel="redirect_uri" href="/redirect">
  </head>
  <body>
    <div class="h-app">
      <img src="/logo.png" class="u-logo">
      <a href="/" class="u-url p-name">Example App</a>
    </div>
  </body>
</html>
`

func TestGet(t *testing.T) {
	t.Parallel()

	client, _, cleanup := util.TestServe(t, func(ctx *http.RequestCtx) {
		ctx.Response.Header.Set(http.HeaderLink, `<https://app.example.net/redirect>; rel="redirect_uri">`)
		ctx.SetStatusCode(http.StatusOK)
		ctx.SetContentType(common.MIMETextHTML)
		ctx.SetBodyString(testBody)
	})
	t.Cleanup(cleanup)

	c := domain.TestClient(t)

	result, err := repository.NewHTTPClientRepository(client).Get(context.TODO(), c.ID)
	require.NoError(t, err)
	assert.Equal(t, c, result)
}
