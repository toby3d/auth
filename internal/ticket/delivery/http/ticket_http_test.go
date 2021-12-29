package http_test

import (
	"fmt"
	"path"
	"sync"
	"testing"

	"github.com/fasthttp/router"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	http "github.com/valyala/fasthttp"

	"source.toby3d.me/website/oauth/internal/common"
	"source.toby3d.me/website/oauth/internal/domain"
	"source.toby3d.me/website/oauth/internal/testing/httptest"
	delivery "source.toby3d.me/website/oauth/internal/ticket/delivery/http"
	ucase "source.toby3d.me/website/oauth/internal/ticket/usecase"
	userrepo "source.toby3d.me/website/oauth/internal/user/repository/memory"
	userucase "source.toby3d.me/website/oauth/internal/user/usecase"
)

// TODO(toby3d): looks ugly, refactor this?
func TestUpdate(t *testing.T) {
	t.Parallel()

	ticket := domain.TestTicket(t)

	// NOTE(toby3d): user token endpoint
	token := domain.TestToken(t)

	store := new(sync.Map)
	store.Store(path.Join(userrepo.DefaultPathPrefix, ticket.Subject.String()), domain.TestUser(t))

	userClient, _, userCleanup := httptest.New(t, func(ctx *http.RequestCtx) {
		ctx.SuccessString(common.MIMEApplicationJSONCharsetUTF8, fmt.Sprintf(`{
			"access_token": "%s",
			"token_type": "Bearer",
			"scope": "%s",
			"me": "%s"
		}`, token.AccessToken, token.Scope.String(), token.Me.String()))
	})
	t.Cleanup(userCleanup)

	// NOTE(toby3d): current token endpoint
	r := router.New()
	client, _, cleanup := httptest.New(t, r.Handler)
	t.Cleanup(cleanup)

	delivery.NewRequestHandler(
		ucase.NewTicketUseCase(userClient), userucase.NewUserUseCase(userrepo.NewMemoryUserRepository(store)),
	).Register(r)

	req := httptest.NewRequest(http.MethodPost, "https://example.com/ticket", []byte(
		`ticket=`+ticket.Ticket+
			`&resource=`+ticket.Resource.String()+
			`&subject=`+ticket.Subject.String(),
	))
	defer http.ReleaseRequest(req)
	req.Header.SetContentType(common.MIMEApplicationForm)

	resp := http.AcquireResponse()
	defer http.ReleaseResponse(resp)

	require.NoError(t, client.Do(req, resp))
	assert.Condition(t, func() bool {
		return resp.StatusCode() == http.StatusOK || resp.StatusCode() == http.StatusAccepted
	}, "the ticket endpoint MUST return an HTTP 200 OK code or HTTP 202 Accepted")
	// TODO(toby3d): print the result as part of the debugging. Instead, you
	// need to send or save the token to the recipient for later use.
	assert.NotNil(t, resp.Body())
}