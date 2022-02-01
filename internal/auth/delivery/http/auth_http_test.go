package http_test

import (
	"path"
	"strings"
	"sync"
	"testing"

	"github.com/fasthttp/router"
	http "github.com/valyala/fasthttp"
	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"source.toby3d.me/website/indieauth/internal/auth"
	delivery "source.toby3d.me/website/indieauth/internal/auth/delivery/http"
	ucase "source.toby3d.me/website/indieauth/internal/auth/usecase"
	"source.toby3d.me/website/indieauth/internal/client"
	clientrepo "source.toby3d.me/website/indieauth/internal/client/repository/memory"
	clientucase "source.toby3d.me/website/indieauth/internal/client/usecase"
	"source.toby3d.me/website/indieauth/internal/domain"
	profilerepo "source.toby3d.me/website/indieauth/internal/profile/repository/memory"
	"source.toby3d.me/website/indieauth/internal/session"
	sessionrepo "source.toby3d.me/website/indieauth/internal/session/repository/memory"
	"source.toby3d.me/website/indieauth/internal/testing/httptest"
	userrepo "source.toby3d.me/website/indieauth/internal/user/repository/memory"
)

type dependencies struct {
	authService   auth.UseCase
	clients       client.Repository
	clientService client.UseCase
	config        *domain.Config
	matcher       language.Matcher
	sessions      session.Repository
	store         *sync.Map
}

func TestRender(t *testing.T) {
	t.Parallel()

	deps := NewDependencies(t)
	me := domain.TestMe(t, "https://user.example.net")
	user := domain.TestUser(t)
	client := domain.TestClient(t)

	deps.store.Store(path.Join(clientrepo.DefaultPathPrefix, client.ID.String()), client)
	deps.store.Store(path.Join(profilerepo.DefaultPathPrefix, me.String()), user.Profile)
	deps.store.Store(path.Join(userrepo.DefaultPathPrefix, me.String()), user)

	r := router.New()
	delivery.NewRequestHandler(delivery.NewRequestHandlerOptions{
		Auth:    deps.authService,
		Clients: deps.clientService,
		Config:  deps.config,
		Matcher: deps.matcher,
	}).Register(r)

	httpClient, _, cleanup := httptest.New(t, r.Handler)
	t.Cleanup(cleanup)

	uri := http.AcquireURI()
	defer http.ReleaseURI(uri)
	uri.Update("https://example.com/authorize")

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
		uri.QueryArgs().Set(key, val)
	}

	req := httptest.NewRequest(http.MethodGet, uri.String(), nil)
	defer http.ReleaseRequest(req)

	resp := http.AcquireResponse()
	defer http.ReleaseResponse(resp)

	if err := httpClient.Do(req, resp); err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode() != http.StatusOK {
		t.Errorf("GET %s = %d, want %d", uri.String(), resp.StatusCode(), http.StatusOK)
	}

	const expResult = `Authorize application`
	if result := string(resp.Body()); !strings.Contains(result, expResult) {
		t.Errorf("GET %s = %s, want %s", uri.String(), result, expResult)
	}
}

func NewDependencies(tb testing.TB) dependencies {
	tb.Helper()

	config := domain.TestConfig(tb)
	matcher := language.NewMatcher(message.DefaultCatalog.Languages())
	store := new(sync.Map)
	clients := clientrepo.NewMemoryClientRepository(store)
	sessions := sessionrepo.NewMemorySessionRepository(config, store)
	authService := ucase.NewAuthUseCase(sessions, config)
	clientService := clientucase.NewClientUseCase(clients)

	return dependencies{
		authService:   authService,
		clients:       clients,
		clientService: clientService,
		config:        config,
		matcher:       matcher,
		sessions:      sessions,
		store:         store,
	}
}
