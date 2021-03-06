package http_test

import (
	"testing"

	"github.com/fasthttp/router"
	http "github.com/valyala/fasthttp"

	delivery "source.toby3d.me/toby3d/auth/internal/health/delivery/http"
	"source.toby3d.me/toby3d/auth/internal/testing/httptest"
)

func TestRequestHandler(t *testing.T) {
	t.Parallel()

	r := router.New()
	delivery.NewRequestHandler().Register(r)

	client, _, cleanup := httptest.New(t, r.Handler)
	t.Cleanup(cleanup)

	const requestURL = "https://app.example.com/health"
	req, resp := httptest.NewRequest(http.MethodGet, requestURL, nil), http.AcquireResponse()

	t.Cleanup(func() {
		http.ReleaseRequest(req)
		http.ReleaseResponse(resp)
	})

	if err := client.Do(req, resp); err != nil {
		t.Fatal(err)
	}

	if result := resp.StatusCode(); result != http.StatusOK {
		t.Errorf("GET %s = %d, want %d", requestURL, result, http.StatusOK)
	}

	const expBody = `{"ok": true}`
	if result := string(resp.Body()); result != expBody {
		t.Errorf("GET %s = %s, want %s", requestURL, result, expBody)
	}
}
