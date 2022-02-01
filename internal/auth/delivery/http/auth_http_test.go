package http_test

import (
	"path"
	"sync"
	"testing"

	"github.com/fasthttp/router"
	"github.com/fasthttp/session/v2"
	"github.com/fasthttp/session/v2/providers/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	http "github.com/valyala/fasthttp"
	"golang.org/x/text/language"
	"golang.org/x/text/message"

	delivery "source.toby3d.me/website/indieauth/internal/auth/delivery/http"
	ucase "source.toby3d.me/website/indieauth/internal/auth/usecase"
	clientrepo "source.toby3d.me/website/indieauth/internal/client/repository/memory"
	clientucase "source.toby3d.me/website/indieauth/internal/client/usecase"
	"source.toby3d.me/website/indieauth/internal/domain"
	profilerepo "source.toby3d.me/website/indieauth/internal/profile/repository/memory"
	sessionrepo "source.toby3d.me/website/indieauth/internal/session/repository/memory"
	"source.toby3d.me/website/indieauth/internal/testing/httptest"
	userrepo "source.toby3d.me/website/indieauth/internal/user/repository/memory"
)

func TestRender(t *testing.T) {
	t.Parallel()

	provider, err := memory.New(memory.Config{})
	require.NoError(t, err)

	s := session.New(session.NewDefaultConfig())
	require.NoError(t, s.SetProvider(provider))

	me := domain.TestMe(t, "https://user.example.net")
	client := domain.TestClient(t)
	config := domain.TestConfig(t)
	store := new(sync.Map)
	user := domain.TestUser(t)
	store.Store(path.Join(userrepo.DefaultPathPrefix, me.String()), user)
	store.Store(path.Join(clientrepo.DefaultPathPrefix, client.ID.String()), client)
	store.Store(path.Join(profilerepo.DefaultPathPrefix, me.String()), user.Profile)

	router := router.New()
	delivery.NewRequestHandler(delivery.NewRequestHandlerOptions{
		Clients: clientucase.NewClientUseCase(clientrepo.NewMemoryClientRepository(store)),
		Config:  config,
		Matcher: language.NewMatcher(message.DefaultCatalog.Languages()),
		Auth: ucase.NewAuthUseCase(
			sessionrepo.NewMemorySessionRepository(config, store),
			config,
		),
	}).Register(router)

	httpClient, _, cleanup := httptest.New(t, router.Handler)
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

	require.NoError(t, httpClient.Do(req, resp))

	assert.Equal(t, http.StatusOK, resp.StatusCode())
	assert.Contains(t, string(resp.Body()), `Authorize application`)
}
