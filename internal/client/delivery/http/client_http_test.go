package http_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"golang.org/x/text/language"
	"golang.org/x/text/message"

	delivery "source.toby3d.me/toby3d/auth/internal/client/delivery/http"
	"source.toby3d.me/toby3d/auth/internal/domain"
	"source.toby3d.me/toby3d/auth/internal/profile"
	profilerepo "source.toby3d.me/toby3d/auth/internal/profile/repository/memory"
	"source.toby3d.me/toby3d/auth/internal/session"
	sessionrepo "source.toby3d.me/toby3d/auth/internal/session/repository/memory"
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
	tokens       token.Repository
	tokenService token.UseCase
}

func TestRead(t *testing.T) {
	t.Parallel()

	deps := NewDependencies(t)
	req := httptest.NewRequest(http.MethodGet, "https://app.example.com/", nil)
	w := httptest.NewRecorder()

	delivery.NewHandler(delivery.NewHandlerOptions{
		Client:  *deps.client,
		Config:  *deps.config,
		Matcher: deps.matcher,
		Tokens:  deps.tokenService,
	}).ServeHTTP(w, req)

	if resp := w.Result(); resp.StatusCode != http.StatusOK {
		t.Errorf("%s %s = %d, want %d", req.Method, req.RequestURI, resp.StatusCode, http.StatusOK)
	}
}

func NewDependencies(tb testing.TB) Dependencies {
	tb.Helper()

	client := domain.TestClient(tb)
	config := domain.TestConfig(tb)
	matcher := language.NewMatcher(message.DefaultCatalog.Languages())
	sessions := sessionrepo.NewMemorySessionRepository(*config)
	tokens := tokenrepo.NewMemoryTokenRepository()
	profiles := profilerepo.NewMemoryProfileRepository()
	tokenService := tokenucase.NewTokenUseCase(tokenucase.Config{
		Config:   *config,
		Profiles: profiles,
		Sessions: sessions,
		Tokens:   tokens,
	})

	return Dependencies{
		client:       client,
		config:       config,
		matcher:      matcher,
		sessions:     sessions,
		profiles:     profiles,
		tokens:       tokens,
		tokenService: tokenService,
	}
}
