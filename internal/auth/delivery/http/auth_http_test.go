package http_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"source.toby3d.me/toby3d/auth/internal/auth"
	delivery "source.toby3d.me/toby3d/auth/internal/auth/delivery/http"
	ucase "source.toby3d.me/toby3d/auth/internal/auth/usecase"
	"source.toby3d.me/toby3d/auth/internal/client"
	clientrepo "source.toby3d.me/toby3d/auth/internal/client/repository/memory"
	clientucase "source.toby3d.me/toby3d/auth/internal/client/usecase"
	"source.toby3d.me/toby3d/auth/internal/domain"
	"source.toby3d.me/toby3d/auth/internal/profile"
	profilerepo "source.toby3d.me/toby3d/auth/internal/profile/repository/memory"
	"source.toby3d.me/toby3d/auth/internal/session"
	sessionrepo "source.toby3d.me/toby3d/auth/internal/session/repository/memory"
	"source.toby3d.me/toby3d/auth/internal/user"
	userrepo "source.toby3d.me/toby3d/auth/internal/user/repository/memory"
)

type Dependencies struct {
	authService   auth.UseCase
	clients       client.Repository
	clientService client.UseCase
	matcher       language.Matcher
	profiles      profile.Repository
	sessions      session.Repository
	users         user.Repository
	config        *domain.Config
}

//nolint:funlen
func TestAuthorize(t *testing.T) {
	t.Parallel()

	deps := NewDependencies(t)
	me := domain.TestMe(t, "https://user.example.net/")
	user := domain.TestUser(t)
	client := domain.TestClient(t)

	if err := deps.clients.Create(context.Background(), *client); err != nil {
		t.Fatal(err)
	}

	if err := deps.users.Create(context.Background(), *user); err != nil {
		t.Fatal(err)
	}

	if err := deps.profiles.Create(context.Background(), *me, *user.Profile); err != nil {
		t.Fatal(err)
	}

	u := &url.URL{Scheme: "https", Host: "example.com", Path: "/"}
	q := u.Query()

	for key, val := range map[string]string{
		"client_id":             client.ID.String(),
		"code_challenge":        "OfYAxt8zU2dAPDWQxTAUIteRzMsoj9QBdMIVEDOErUo",
		"code_challenge_method": domain.CodeChallengeMethodS256.String(),
		"me":                    me.String(),
		"redirect_uri":          client.RedirectURI[0].String(),
		"response_type":         domain.ResponseTypeCode.String(),
		"scope":                 "profile email",
		"state":                 "1234567890",
	} {
		q.Set(key, val)
	}

	u.RawQuery = q.Encode()

	req := httptest.NewRequest(http.MethodGet, u.String(), nil)
	w := httptest.NewRecorder()

	//nolint:exhaustivestruct
	delivery.NewHandler(delivery.NewHandlerOptions{
		Auth:    deps.authService,
		Clients: deps.clientService,
		Config:  *deps.config,
		Matcher: deps.matcher,
	}).ServeHTTP(w, req)

	resp := w.Result()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("%s %s = %d, want %d", req.Method, u.String(), resp.StatusCode, http.StatusOK)
	}

	expResult := `Authorize ` + client.Name
	if result := string(body); !strings.Contains(result, expResult) {
		t.Errorf("%s %s = %s, want %s", req.Method, u.String(), result, expResult)
	}
}

func NewDependencies(tb testing.TB) Dependencies {
	tb.Helper()

	config := domain.TestConfig(tb)
	matcher := language.NewMatcher(message.DefaultCatalog.Languages())
	clients := clientrepo.NewMemoryClientRepository()
	users := userrepo.NewMemoryUserRepository()
	sessions := sessionrepo.NewMemorySessionRepository(*config)
	profiles := profilerepo.NewMemoryProfileRepository()
	authService := ucase.NewAuthUseCase(sessions, profiles, *config)
	clientService := clientucase.NewClientUseCase(clients)

	return Dependencies{
		users:         users,
		authService:   authService,
		clients:       clients,
		clientService: clientService,
		config:        config,
		matcher:       matcher,
		sessions:      sessions,
		profiles:      profiles,
	}
}
