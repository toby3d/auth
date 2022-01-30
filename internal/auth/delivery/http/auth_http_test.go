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
	c := domain.TestClient(t)
	config := domain.TestConfig(t)
	store := new(sync.Map)
	user := domain.TestUser(t)
	store.Store(path.Join(userrepo.DefaultPathPrefix, me.String()), user)
	store.Store(path.Join(clientrepo.DefaultPathPrefix, c.ID.String()), c)
	store.Store(path.Join(profilerepo.DefaultPathPrefix, me.String()), user.Profile)

	r := router.New()
	delivery.NewRequestHandler(delivery.NewRequestHandlerOptions{
		Clients: clientucase.NewClientUseCase(clientrepo.NewMemoryClientRepository(store)),
		Config:  config,
		Matcher: language.NewMatcher(message.DefaultCatalog.Languages()),
		Auth: ucase.NewAuthUseCase(
			sessionrepo.NewMemorySessionRepository(config, store),
			config,
		),
	}).Register(r)

	client, _, cleanup := httptest.New(t, r.Handler)
	t.Cleanup(cleanup)

	u := http.AcquireURI()
	defer http.ReleaseURI(u)
	u.Update("https://example.com/authorize")

	for k, v := range map[string]string{
		"client_id":             c.ID.String(),
		"code_challenge":        "OfYAxt8zU2dAPDWQxTAUIteRzMsoj9QBdMIVEDOErUo",
		"code_challenge_method": domain.CodeChallengeMethodS256.String(),
		"me":                    me.String(),
		"redirect_uri":          c.RedirectURI[0].String(),
		"response_type":         domain.ResponseTypeCode.String(),
		"scope":                 "profile email",
		"state":                 "1234567890",
	} {
		u.QueryArgs().Set(k, v)
	}

	req := httptest.NewRequest(http.MethodGet, u.String(), nil)
	defer http.ReleaseRequest(req)

	resp := http.AcquireResponse()
	defer http.ReleaseResponse(resp)

	require.NoError(t, client.Do(req, resp))

	assert.Equal(t, http.StatusOK, resp.StatusCode())
	assert.Contains(t, string(resp.Body()), `Authorize application`)
}
