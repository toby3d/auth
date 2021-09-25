package http_test

import (
	"context"
	"net"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	http "github.com/valyala/fasthttp"
	httputil "github.com/valyala/fasthttp/fasthttputil"
	repository "source.toby3d.me/website/oauth/internal/client/repository/http"
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

	ln := httputil.NewInmemoryListener()
	u := http.AcquireURI()
	u.SetScheme("http")
	u.SetHost(ln.Addr().String())

	t.Cleanup(func() {
		http.ReleaseURI(u)
		assert.NoError(t, ln.Close())
	})

	go func(t *testing.T) {
		t.Helper()
		require.NoError(t, http.Serve(ln, func(ctx *http.RequestCtx) {
			ctx.SuccessString(common.MIMETextHTML, testBody)
			ctx.Response.Header.Set(http.HeaderLink,
				`<https://app.example.com/redirect>; rel="redirect_uri">`)
		}))
	}(t)

	client := new(http.Client)
	client.Dial = func(addr string) (net.Conn, error) {
		conn, err := ln.Dial()
		if err != nil {
			return nil, errors.Wrap(err, "failed to dial the address")
		}

		return conn, nil
	}

	result, err := repository.NewHTTPClientRepository(client).Get(context.TODO(), u.String())
	require.NoError(t, err)
	assert.Equal(t, &model.Client{
		ID:   model.URL(u.String()),
		Name: "Example App",
		Logo: model.URL(u.String() + "logo.png"),
		URL:  model.URL(u.String()),
		RedirectURI: []model.URL{
			"https://app.example.com/redirect",
			model.URL(u.String() + "redirect"),
		},
	}, result)
}
