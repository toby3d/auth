package http_test

import (
	"sync"
	"testing"

	"github.com/fasthttp/router"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	http "github.com/valyala/fasthttp"
	"golang.org/x/text/language"
	"golang.org/x/text/message"

	delivery "source.toby3d.me/website/indieauth/internal/client/delivery/http"
	"source.toby3d.me/website/indieauth/internal/domain"
	sessionrepo "source.toby3d.me/website/indieauth/internal/session/repository/memory"
	"source.toby3d.me/website/indieauth/internal/testing/httptest"
	tokenrepo "source.toby3d.me/website/indieauth/internal/token/repository/memory"
	tokenucase "source.toby3d.me/website/indieauth/internal/token/usecase"
)

func TestRead(t *testing.T) {
	t.Parallel()

	store := new(sync.Map)
	config := domain.TestConfig(t)

	r := router.New()
	delivery.NewRequestHandler(delivery.NewRequestHandlerOptions{
		Client:  domain.TestClient(t),
		Config:  config,
		Matcher: language.NewMatcher(message.DefaultCatalog.Languages()),
		Tokens: tokenucase.NewTokenUseCase(tokenrepo.NewMemoryTokenRepository(store),
			sessionrepo.NewMemorySessionRepository(config, store), config),
	}).Register(r)

	client, _, cleanup := httptest.New(t, r.Handler)
	t.Cleanup(cleanup)

	req := httptest.NewRequest(http.MethodGet, "https://app.example.com/", nil)
	defer http.ReleaseRequest(req)

	resp := http.AcquireResponse()
	defer http.ReleaseResponse(resp)

	require.NoError(t, client.Do(req, resp))
	assert.Equal(t, http.StatusOK, resp.StatusCode())
}
