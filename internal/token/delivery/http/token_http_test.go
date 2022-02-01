package http_test

import (
	"context"
	"strings"
	"sync"
	"testing"

	"github.com/fasthttp/router"
	json "github.com/goccy/go-json"
	"github.com/stretchr/testify/assert"
	http "github.com/valyala/fasthttp"

	"source.toby3d.me/website/indieauth/internal/common"
	"source.toby3d.me/website/indieauth/internal/domain"
	sessionrepo "source.toby3d.me/website/indieauth/internal/session/repository/memory"
	"source.toby3d.me/website/indieauth/internal/testing/httptest"
	ticketrepo "source.toby3d.me/website/indieauth/internal/ticket/repository/memory"
	ticketucase "source.toby3d.me/website/indieauth/internal/ticket/usecase"
	delivery "source.toby3d.me/website/indieauth/internal/token/delivery/http"
	tokenrepo "source.toby3d.me/website/indieauth/internal/token/repository/memory"
	tokenucase "source.toby3d.me/website/indieauth/internal/token/usecase"
)

/* TODO(toby3d)
func TestExchange(t *testing.T) {
	t.Parallel()
}
*/

func TestVerification(t *testing.T) {
	t.Parallel()

	store := new(sync.Map)
	config := domain.TestConfig(t)
	token := domain.TestToken(t)

	router := router.New()
	// TODO(toby3d): provide tickets
	delivery.NewRequestHandler(
		tokenucase.NewTokenUseCase(
			tokenrepo.NewMemoryTokenRepository(store),
			sessionrepo.NewMemorySessionRepository(config, store),
			config,
		),
		ticketucase.NewTicketUseCase(
			ticketrepo.NewMemoryTicketRepository(store, config),
			new(http.Client),
			config,
		),
	).Register(router)

	client, _, cleanup := httptest.New(t, router.Handler)
	t.Cleanup(cleanup)

	req := httptest.NewRequest(http.MethodGet, "https://app.example.com/token", nil)
	defer http.ReleaseRequest(req)
	req.Header.Set(http.HeaderAccept, common.MIMEApplicationJSON)
	token.SetAuthHeader(req)

	resp := http.AcquireResponse()
	defer http.ReleaseResponse(resp)

	if err := client.Do(req, resp); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, http.StatusOK, resp.StatusCode())

	result := new(delivery.TokenVerificationResponse)
	if err := json.Unmarshal(resp.Body(), result); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, token.ClientID.String(), result.ClientID.String())
	assert.Equal(t, token.Me.String(), result.Me.String())
	assert.Equal(t, token.Scope.String(), result.Scope.String())
}

func TestRevocation(t *testing.T) {
	t.Parallel()

	config := domain.TestConfig(t)
	store := new(sync.Map)
	tokens := tokenrepo.NewMemoryTokenRepository(store)
	accessToken := domain.TestToken(t)

	router := router.New()
	delivery.NewRequestHandler(
		tokenucase.NewTokenUseCase(
			tokens,
			sessionrepo.NewMemorySessionRepository(config, store),
			config,
		),
		ticketucase.NewTicketUseCase(
			ticketrepo.NewMemoryTicketRepository(store, config),
			new(http.Client),
			config,
		),
	).Register(router)

	client, _, cleanup := httptest.New(t, router.Handler)
	t.Cleanup(cleanup)

	req := httptest.NewRequest(http.MethodPost, "https://app.example.com/token", nil)
	defer http.ReleaseRequest(req)
	req.Header.Set(http.HeaderAccept, common.MIMEApplicationJSON)
	req.Header.SetContentType(common.MIMEApplicationForm)
	req.PostArgs().Set("action", domain.ActionRevoke.String())
	req.PostArgs().Set("token", accessToken.AccessToken)

	resp := http.AcquireResponse()
	defer http.ReleaseResponse(resp)

	if err := client.Do(req, resp); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, http.StatusOK, resp.StatusCode())
	assert.Equal(t, `{}`, strings.TrimSpace(string(resp.Body())))

	result, err := tokens.Get(context.TODO(), accessToken.AccessToken)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, accessToken.AccessToken, result.AccessToken)
}
