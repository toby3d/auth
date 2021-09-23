package http_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
	"source.toby3d.me/website/oauth/internal/client/repository/http"
	"source.toby3d.me/website/oauth/internal/common"
	"source.toby3d.me/website/oauth/internal/model"
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

	//nolint: exhaustivestruct
	srv := &fasthttp.Server{
		ReduceMemoryUsage: true,
		GetOnly:           true,
		CloseOnShutdown:   true,
		Handler: func(ctx *fasthttp.RequestCtx) {
			ctx.SuccessString(common.MIMETextHTML, testBody)
			ctx.Response.Header.Set(fasthttp.HeaderLink, `<https://app.example.com/redirect>; rel="redirect_uri">`)
		},
	}

	go func(srv *fasthttp.Server) {
		assert.NoError(t, srv.ListenAndServe("127.0.0.1:2368"))
	}(srv)

	t.Cleanup(func() {
		assert.NoError(t, srv.Shutdown())
	})

	result, err := http.NewHTTPClientRepository(new(fasthttp.Client)).Get(context.TODO(), "http://127.0.0.1:2368/")
	require.NoError(t, err)
	assert.Equal(t, &model.Client{
		ID:   "http://127.0.0.1:2368/",
		Name: "Example App",
		Logo: "http://127.0.0.1:2368/logo.png",
		URL:  "http://127.0.0.1:2368/",
		RedirectURI: []model.URL{
			"https://app.example.com/redirect",
			"http://127.0.0.1:2368/redirect",
		},
	}, result)
}
