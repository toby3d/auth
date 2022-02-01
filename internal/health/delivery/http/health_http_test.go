package http_test

import (
	"testing"

	"github.com/fasthttp/router"
	"github.com/stretchr/testify/assert"
	http "github.com/valyala/fasthttp"

	delivery "source.toby3d.me/website/indieauth/internal/health/delivery/http"
	"source.toby3d.me/website/indieauth/internal/testing/httptest"
)

func TestRequestHandler(t *testing.T) {
	t.Parallel()

	r := router.New()
	delivery.NewRequestHandler().Register(r)

	client, _, cleanup := httptest.New(t, r.Handler)
	t.Cleanup(cleanup)

	req := httptest.NewRequest(http.MethodGet, "https://app.example.com/health", nil)
	defer http.ReleaseRequest(req)

	resp := http.AcquireResponse()
	defer http.ReleaseResponse(resp)

	if err := client.Do(req, resp); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, http.StatusOK, resp.StatusCode())
	assert.Equal(t, `{"ok": true}`, string(resp.Body()))
}
