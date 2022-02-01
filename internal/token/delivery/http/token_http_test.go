package http_test

import (
	"bytes"
	"context"
	"sync"
	"testing"

	"github.com/fasthttp/router"
	json "github.com/goccy/go-json"
	http "github.com/valyala/fasthttp"

	"source.toby3d.me/website/indieauth/internal/common"
	"source.toby3d.me/website/indieauth/internal/domain"
	"source.toby3d.me/website/indieauth/internal/session"
	sessionrepo "source.toby3d.me/website/indieauth/internal/session/repository/memory"
	"source.toby3d.me/website/indieauth/internal/testing/httptest"
	"source.toby3d.me/website/indieauth/internal/ticket"
	ticketrepo "source.toby3d.me/website/indieauth/internal/ticket/repository/memory"
	ticketucase "source.toby3d.me/website/indieauth/internal/ticket/usecase"
	"source.toby3d.me/website/indieauth/internal/token"
	delivery "source.toby3d.me/website/indieauth/internal/token/delivery/http"
	tokenrepo "source.toby3d.me/website/indieauth/internal/token/repository/memory"
	tokenucase "source.toby3d.me/website/indieauth/internal/token/usecase"
)

type Dependencies struct {
	client        *http.Client
	config        *domain.Config
	sessions      session.Repository
	store         *sync.Map
	tickets       ticket.Repository
	ticketService ticket.UseCase
	token         *domain.Token
	tokens        token.Repository
	tokenService  token.UseCase
}

/* TODO(toby3d)
func TestExchange(t *testing.T) {
	t.Parallel()
}
*/

func TestVerification(t *testing.T) {
	t.Parallel()

	deps := NewDependencies(t)

	router := router.New()
	delivery.NewRequestHandler(deps.tokenService, deps.ticketService).Register(router)

	client, _, cleanup := httptest.New(t, router.Handler)
	t.Cleanup(cleanup)

	const requestURL = "https://app.example.com/token"

	req := httptest.NewRequest(http.MethodGet, requestURL, nil)
	defer http.ReleaseRequest(req)
	req.Header.Set(http.HeaderAccept, common.MIMEApplicationJSON)
	deps.token.SetAuthHeader(req)

	resp := http.AcquireResponse()
	defer http.ReleaseResponse(resp)

	if err := client.Do(req, resp); err != nil {
		t.Fatal(err)
	}

	if result := resp.StatusCode(); result != http.StatusOK {
		t.Errorf("GET %s = %d, want %d", requestURL, result, http.StatusOK)
	}

	result := new(delivery.TokenVerificationResponse)
	if err := json.Unmarshal(resp.Body(), result); err != nil {
		t.Fatal(err)
	}

	deps.token.AccessToken = ""

	if result.ClientID.String() != deps.token.ClientID.String() ||
		result.Me.String() != deps.token.Me.String() ||
		result.Scope.String() != deps.token.Scope.String() {
		t.Errorf("GET %s = %+v, want %+v", requestURL, result, deps.token)
	}
}

func TestRevocation(t *testing.T) {
	t.Parallel()

	deps := NewDependencies(t)

	r := router.New()
	delivery.NewRequestHandler(deps.tokenService, deps.ticketService).Register(r)

	client, _, cleanup := httptest.New(t, r.Handler)
	t.Cleanup(cleanup)

	const requestURL = "https://app.example.com/token"

	req := httptest.NewRequest(http.MethodPost, requestURL, nil)
	defer http.ReleaseRequest(req)
	req.Header.Set(http.HeaderAccept, common.MIMEApplicationJSON)
	req.Header.SetContentType(common.MIMEApplicationForm)
	req.PostArgs().Set("action", domain.ActionRevoke.String())
	req.PostArgs().Set("token", deps.token.AccessToken)

	resp := http.AcquireResponse()
	defer http.ReleaseResponse(resp)

	if err := client.Do(req, resp); err != nil {
		t.Fatal(err)
	}

	if result := resp.StatusCode(); result != http.StatusOK {
		t.Errorf("POST %s = %d, want %d", requestURL, result, http.StatusOK)
	}

	expBody := []byte("{}") //nolint: ifshort
	if result := bytes.TrimSpace(resp.Body()); !bytes.Equal(result, expBody) {
		t.Errorf("POST %s = %s, want %s", requestURL, result, expBody)
	}

	result, err := deps.tokens.Get(context.TODO(), deps.token.AccessToken)
	if err != nil {
		t.Fatal(err)
	}

	if result.String() != deps.token.String() {
		t.Errorf("Get(%+v) = %s, want %s", deps.token.AccessToken, result, deps.token)
	}
}

func NewDependencies(tb testing.TB) Dependencies {
	tb.Helper()

	client := new(http.Client)
	config := domain.TestConfig(tb)
	store := new(sync.Map)
	token := domain.TestToken(tb)
	sessions := sessionrepo.NewMemorySessionRepository(config, store)
	tickets := ticketrepo.NewMemoryTicketRepository(store, config)
	tokens := tokenrepo.NewMemoryTokenRepository(store)
	ticketService := ticketucase.NewTicketUseCase(tickets, client, config)
	tokenService := tokenucase.NewTokenUseCase(tokens, sessions, config)

	return Dependencies{
		client:        client,
		config:        config,
		sessions:      sessions,
		store:         store,
		tickets:       tickets,
		ticketService: ticketService,
		token:         token,
		tokens:        tokens,
		tokenService:  tokenService,
	}
}
