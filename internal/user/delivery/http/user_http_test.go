package http_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/goccy/go-json"
	"github.com/google/go-cmp/cmp"

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
	token        *domain.Token
	tokens       token.Repository
	tokenService token.UseCase
}

func TestUserInfo(t *testing.T) {
	t.Parallel()

	deps := NewDependencies(t)
	if err := deps.profiles.Create(context.Background(), deps.token.Me, *deps.profile); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "https://example.com/userinfo", nil)
	req.Header.Set(common.HeaderAuthorization, "Bearer "+deps.token.AccessToken)

	w := httptest.NewRecorder()
	delivery.NewHandler(deps.tokenService, deps.config).
		ServeHTTP(w, req)

	resp := w.Result()

	if exp := http.StatusOK; resp.StatusCode != exp {
		t.Errorf("%s %s = %d, want %d", req.Method, req.RequestURI, resp.StatusCode, exp)
	}

	result := new(delivery.UserInformationResponse)
	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		t.Fatal(err)
	}

	exp := &delivery.UserInformationResponse{
		Name:  deps.profile.GetName(),
		URL:   &domain.URL{URL: deps.profile.GetURL()},
		Photo: &domain.URL{URL: deps.profile.GetPhoto()},
		Email: deps.profile.GetEmail(),
	}

	if diff := cmp.Diff(result, exp, cmp.AllowUnexported(domain.URL{}, domain.Email{})); diff != "" {
		t.Errorf("%s %s = %+v, want %+v", req.Method, req.RequestURI, result, exp)
	}
}

func NewDependencies(tb testing.TB) Dependencies {
	tb.Helper()

	config := domain.TestConfig(tb)
	sessions := sessionrepo.NewMemorySessionRepository(*config)
	tokens := tokenrepo.NewMemoryTokenRepository()
	profiles := profilerepo.NewMemoryProfileRepository()

	return Dependencies{
		config:   config,
		profile:  domain.TestProfile(tb),
		profiles: profiles,
		sessions: sessions,
		token:    domain.TestToken(tb),
		tokens:   tokens,
		tokenService: tokenucase.NewTokenUseCase(tokenucase.Config{
			Config:   config,
			Profiles: profiles,
			Sessions: sessions,
			Tokens:   tokens,
		}),
	}
}
