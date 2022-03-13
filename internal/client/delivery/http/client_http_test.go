package http_test

import (
	"sync"
	"testing"

	"github.com/fasthttp/router"
	http "github.com/valyala/fasthttp"
	"golang.org/x/text/language"
	"golang.org/x/text/message"

	delivery "source.toby3d.me/toby3d/auth/internal/client/delivery/http"
	"source.toby3d.me/toby3d/auth/internal/domain"
	"source.toby3d.me/toby3d/auth/internal/profile"
	profilerepo "source.toby3d.me/toby3d/auth/internal/profile/repository/memory"
	"source.toby3d.me/toby3d/auth/internal/session"
	sessionrepo "source.toby3d.me/toby3d/auth/internal/session/repository/memory"
	"source.toby3d.me/toby3d/auth/internal/testing/httptest"
	"source.toby3d.me/toby3d/auth/internal/token"
	tokenrepo "source.toby3d.me/toby3d/auth/internal/token/repository/memory"
	tokenucase "source.toby3d.me/toby3d/auth/internal/token/usecase"
)

type Dependencies struct {
	profiles     profile.Repository
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

	r := router.New() //nolint: varnamelen
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

func NewDependencies(tb testing.TB) Dependencies {
	tb.Helper()

	store := new(sync.Map)
	client := domain.TestClient(tb)
	config := domain.TestConfig(tb)
	matcher := language.NewMatcher(message.DefaultCatalog.Languages())
	sessions := sessionrepo.NewMemorySessionRepository(store, config)
	tokens := tokenrepo.NewMemoryTokenRepository(store)
	profiles := profilerepo.NewMemoryProfileRepository(store)
	tokenService := tokenucase.NewTokenUseCase(tokenucase.Config{
		Config:   config,
		Profiles: profiles,
		Sessions: sessions,
		Tokens:   tokens,
	})

	return Dependencies{
		client:       client,
		config:       config,
		matcher:      matcher,
		sessions:     sessions,
		store:        store,
		profiles:     profiles,
		tokens:       tokens,
		tokenService: tokenService,
	}
}
