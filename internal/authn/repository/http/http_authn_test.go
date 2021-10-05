package http_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	http "github.com/valyala/fasthttp"

	repository "source.toby3d.me/website/oauth/internal/authn/repository/http"
	"source.toby3d.me/website/oauth/internal/common"
	"source.toby3d.me/website/oauth/internal/util"
)

const testBody string = `
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Example Profile</title>
    <link rel="me authn" href="https://user.example.com/">
  </head>
  <body>
    <div class="h-card">
      <img src="/photo.png" class="u-photo">
      <a rel="me" href="https://user.example.org/">Unsecure profile</a>
      <a rel="me authn" href="https://user.example.net/">Secure profile</a>
    </div>
  </body>
</html>
`

func TestFetch(t *testing.T) {
	t.Parallel()

	client, _, cleanup := util.TestServe(t, func(ctx *http.RequestCtx) {
		ctx.SetStatusCode(http.StatusOK)
		ctx.SetContentType(common.MIMETextHTML)
		ctx.SetBodyString(testBody)
	})
	t.Cleanup(cleanup)

	result, err := repository.NewHTTPAuthnRepository(client).Fetch(context.TODO(), "https://example.com/")
	assert.NoError(t, err)
	assert.Equal(t, []string{
		"https://user.example.com/",
		"https://user.example.net/",
	}, result)
}
