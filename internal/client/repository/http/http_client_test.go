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

func TestGet(t *testing.T) {
	t.Parallel()

	httpClient, _, cleanup := util.TestServe(t, func(ctx *http.RequestCtx) {
		ctx.SuccessString(common.MIMETextHTML, `
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
		`)
		ctx.Response.Header.Set(http.HeaderLink, `<http://app.example.com/redirect>; rel="redirect_uri">`)
	})
	t.Cleanup(cleanup)

	client := domain.TestClient(t)

	result, err := repository.NewHTTPClientRepository(httpClient).Get(context.TODO(), client.ID)
	require.NoError(t, err)
	assert.Equal(t, client, result)
}
