package http_test

import (
	"io"
	"net/http/httptest"
	"testing"

	http "github.com/valyala/fasthttp"

	delivery "source.toby3d.me/toby3d/auth/internal/health/delivery/http"
)

func TestRequestHandler(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "https://example.com/health", nil)
	w := httptest.NewRecorder()
	delivery.NewHandler().ServeHTTP(w, req)

	resp := w.Result()

	if exp := http.StatusOK; resp.StatusCode != exp {
		t.Errorf("%s %s = %d, want %d", req.Method, req.RequestURI, resp.StatusCode, exp)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	if exp := `👌`; string(body) != exp {
		t.Errorf("%s %s = '%s', want '%s'", req.Method, req.RequestURI, body, exp)
	}
}
