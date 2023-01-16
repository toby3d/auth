package http_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/goccy/go-json"

	"source.toby3d.me/toby3d/auth/internal/domain"
	delivery "source.toby3d.me/toby3d/auth/internal/metadata/delivery/http"
)

func TestMetadata(t *testing.T) {
	t.Parallel()

	metadata := domain.TestMetadata(t)

	req := httptest.NewRequest(http.MethodGet, "https://example.com/.well-known/oauth-authorization-server", nil)

	w := httptest.NewRecorder()
	delivery.NewHandler(metadata).
		Handler().
		ServeHTTP(w, req)

	resp := w.Result()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("%s %s = %d, want %d", req.Method, req.RequestURI, resp.StatusCode, http.StatusOK)
	}

	out := new(delivery.MetadataResponse)
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		t.Fatal(err)
	}
}
