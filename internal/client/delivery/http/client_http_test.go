package http_test

import (
	"sync"
	"testing"

	"github.com/fasthttp/router"
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

	router := router.New()
	delivery.NewRequestHandler(delivery.NewRequestHandlerOptions{
		Client:  domain.TestClient(t),
		Config:  config,
		Matcher: language.NewMatcher(message.DefaultCatalog.Languages()),
		Tokens: tokenucase.NewTokenUseCase(tokenrepo.NewMemoryTokenRepository(store),
			sessionrepo.NewMemorySessionRepository(config, store), config),
	}).Register(router)

	client, _, cleanup := httptest.New(t, router.Handler)
	t.Cleanup(cleanup)

	const requestURI string = "https://app.example.com/"

	req := httptest.NewRequest(http.MethodGet, requestURI, nil)
	defer http.ReleaseRequest(req)

	resp := http.AcquireResponse()
	defer http.ReleaseResponse(resp)

	if err := client.Do(req, resp); err != nil {
		t.Error(err)
	}

	if resp.StatusCode() != http.StatusOK {
		t.Errorf("GET %s = %d, want %d", requestURI, resp.StatusCode(), http.StatusOK)
	}
}
