package http_test

import (
	"net/http/httptest"
	"path"
	"sync"
	"testing"

	"github.com/goccy/go-json"
	http "github.com/valyala/fasthttp"

	"source.toby3d.me/toby3d/auth/internal/common"
	"source.toby3d.me/toby3d/auth/internal/domain"
	"source.toby3d.me/toby3d/auth/internal/profile"
	profilerepo "source.toby3d.me/toby3d/auth/internal/profile/repository/memory"
	"source.toby3d.me/toby3d/auth/internal/session"
	sessionrepo "source.toby3d.me/toby3d/auth/internal/session/repository/memory"
	"source.toby3d.me/toby3d/auth/internal/token"
	tokenrepo "source.toby3d.me/toby3d/auth/internal/token/repository/memory"
	tokenucase "source.toby3d.me/toby3d/auth/internal/token/usecase"
	delivery "source.toby3d.me/toby3d/auth/internal/user/delivery/http"
)

type Dependencies struct {
	config       *domain.Config
	profile      *domain.Profile
	profiles     profile.Repository
	sessions     session.Repository
	store        *sync.Map
	token        *domain.Token
	tokens       token.Repository
	tokenService token.UseCase
}

func TestUserInfo(t *testing.T) {
	t.Parallel()

	deps := NewDependencies(t)
	deps.store.Store(path.Join(profilerepo.DefaultPathPrefix, deps.token.Me.String()), deps.profile)

	req := httptest.NewRequest(http.MethodGet, "https://example.com/userinfo", nil)
	req.Header.Set(common.HeaderAuthorization, "Bearer "+deps.token.AccessToken)

	w := httptest.NewRecorder()
	delivery.NewHandler(deps.tokenService, deps.config).ServeHTTP(w, req)

	resp := w.Result()

	if exp := http.StatusOK; resp.StatusCode != exp {
		t.Errorf("%s %s = %d, want %d", req.Method, req.RequestURI, resp.StatusCode, exp)
	}

	result := new(delivery.UserInformationResponse)
	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		t.Fatal(err)
	}

	if result.Name != deps.profile.GetName() ||
		result.Photo != deps.profile.GetPhoto().String() {
		t.Errorf("GET /userinfo = %+v, want %+v", result, &delivery.UserInformationResponse{
			Name:  deps.profile.GetName(),
			URL:   deps.profile.GetURL().String(),
			Photo: deps.profile.GetPhoto().String(),
			Email: deps.profile.GetEmail().String(),
		})
	}
}

func NewDependencies(tb testing.TB) Dependencies {
	tb.Helper()

	store := new(sync.Map)
	config := domain.TestConfig(tb)

	return Dependencies{
		config:   config,
		profile:  domain.TestProfile(tb),
		profiles: profilerepo.NewMemoryProfileRepository(store),
		sessions: sessionrepo.NewMemorySessionRepository(store, config),
		store:    store,
		token:    domain.TestToken(tb),
		tokens:   tokenrepo.NewMemoryTokenRepository(store),
		tokenService: tokenucase.NewTokenUseCase(tokenucase.Config{
			Config:   config,
			Profiles: profilerepo.NewMemoryProfileRepository(store),
			Sessions: sessionrepo.NewMemorySessionRepository(store, config),
			Tokens:   tokenrepo.NewMemoryTokenRepository(store),
		}),
	}
}
