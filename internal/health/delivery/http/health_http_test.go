package http_test

import (
	"testing"

	"github.com/fasthttp/router"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	http "github.com/valyala/fasthttp"

	delivery "source.toby3d.me/website/oauth/internal/health/delivery/http"
	"source.toby3d.me/website/oauth/internal/util"
)

func TestRequestHandler(t *testing.T) {
	t.Parallel()

	r := router.New()
	delivery.NewRequestHandler().Register(r)

	client, _, cleanup := util.TestServe(t, r.Handler)
	t.Cleanup(cleanup)

	req := http.AcquireRequest()
	defer http.ReleaseRequest(req)
	req.Header.SetMethod(http.MethodGet)
	req.SetRequestURI("https://app.example.com/health")

	resp := http.AcquireResponse()
	defer http.ReleaseResponse(resp)

	require.NoError(t, client.Do(req, resp))
	assert.Equal(t, http.StatusOK, resp.StatusCode())
}
