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
	"source.toby3d.me/website/indieauth/internal/session"
	sessionrepo "source.toby3d.me/website/indieauth/internal/session/repository/memory"
	"source.toby3d.me/website/indieauth/internal/testing/httptest"
	"source.toby3d.me/website/indieauth/internal/token"
	tokenrepo "source.toby3d.me/website/indieauth/internal/token/repository/memory"
	tokenucase "source.toby3d.me/website/indieauth/internal/token/usecase"
)

type dependencies struct {
	client       *domain.Client
	config       *domain.Config
	matcher      language.Matcher
	sessions     session.Repository
	store        *sync.Map
	tokens       token.Repository
	tokenService token.UseCase
}

func TestRead(t *testing.T) {
	t.Parallel()

	deps := NewDependencies(t)

	r := router.New()
	delivery.NewRequestHandler(delivery.NewRequestHandlerOptions{
		Client:  deps.client,
		Config:  deps.config,
		Matcher: deps.matcher,
		Tokens:  deps.tokenService,
	}).Register(r)

	client, _, cleanup := httptest.New(t, r.Handler)
	t.Cleanup(cleanup)

	const requestURI string = "https://app.example.com/"
	req, resp := httptest.NewRequest(http.MethodGet, requestURI, nil), http.AcquireResponse()

	t.Cleanup(func() {
		http.ReleaseRequest(req)
		http.ReleaseResponse(resp)
	})

	if err := client.Do(req, resp); err != nil {
		t.Error(err)
	}

	if resp.StatusCode() != http.StatusOK {
		t.Errorf("GET %s = %d, want %d", requestURI, resp.StatusCode(), http.StatusOK)
	}
}

func NewDependencies(tb testing.TB) dependencies {
	tb.Helper()

	client := domain.TestClient(tb)
	config := domain.TestConfig(tb)
	matcher := language.NewMatcher(message.DefaultCatalog.Languages())
	store := new(sync.Map)
	sessions := sessionrepo.NewMemorySessionRepository(config, store)
	tokens := tokenrepo.NewMemoryTokenRepository(store)
	tokenService := tokenucase.NewTokenUseCase(tokens, sessions, config)

	return dependencies{
		client:       client,
		config:       config,
		matcher:      matcher,
		sessions:     sessions,
		store:        store,
		tokens:       tokens,
		tokenService: tokenService,
	}
}
