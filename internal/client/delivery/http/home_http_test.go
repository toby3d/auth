package http_test

import (
	"testing"

	"github.com/fasthttp/router"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	http "github.com/valyala/fasthttp"
	"golang.org/x/text/language"
	"golang.org/x/text/message"

	delivery "source.toby3d.me/website/indieauth/internal/client/delivery/http"
	"source.toby3d.me/website/indieauth/internal/domain"
	"source.toby3d.me/website/indieauth/internal/testing/httptest"
)

func TestRead(t *testing.T) {
	t.Parallel()

	r := router.New()
	delivery.NewRequestHandler(
		domain.TestConfig(t), domain.TestClient(t), language.NewMatcher(message.DefaultCatalog.Languages()),
	).Register(r)

	client, _, cleanup := httptest.New(t, r.Handler)
	t.Cleanup(cleanup)

	req := httptest.NewRequest(http.MethodGet, "https://app.example.com/", nil)
	defer http.ReleaseRequest(req)

	resp := http.AcquireResponse()
	defer http.ReleaseResponse(resp)

	require.NoError(t, client.Do(req, resp))
	assert.Equal(t, http.StatusOK, resp.StatusCode())
}
