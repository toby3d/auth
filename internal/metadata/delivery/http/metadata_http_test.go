package http_test

import (
	"testing"

	"github.com/fasthttp/router"
	"github.com/goccy/go-json"
	http "github.com/valyala/fasthttp"

	"source.toby3d.me/toby3d/auth/internal/domain"
	delivery "source.toby3d.me/toby3d/auth/internal/metadata/delivery/http"
	"source.toby3d.me/toby3d/auth/internal/testing/httptest"
)

func TestMetadata(t *testing.T) {
	t.Parallel()

	r := router.New()
	metadata := domain.TestMetadata(t)
	delivery.NewRequestHandler(metadata).Register(r)

	client, _, cleanup := httptest.New(t, r.Handler)
	t.Cleanup(cleanup)

	const requestURL string = "https://example.com/.well-known/oauth-authorization-server"

	status, body, err := client.Get(nil, requestURL)
	if err != nil {
		t.Fatal(err)
	}

	if status != http.StatusOK {
		t.Errorf("GET %s = %d, want %d", requestURL, status, http.StatusOK)
	}

	result := new(delivery.MetadataResponse)
	if err = json.Unmarshal(body, result); err != nil {
		t.Fatal(err)
	}
}
