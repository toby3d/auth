package http_test

import (
	"sync"
	"testing"

	"github.com/fasthttp/router"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	http "github.com/valyala/fasthttp"
	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"source.toby3d.me/website/indieauth/internal/common"
	"source.toby3d.me/website/indieauth/internal/domain"
	"source.toby3d.me/website/indieauth/internal/testing/httptest"
	delivery "source.toby3d.me/website/indieauth/internal/ticket/delivery/http"
	ticketrepo "source.toby3d.me/website/indieauth/internal/ticket/repository/memory"
	ucase "source.toby3d.me/website/indieauth/internal/ticket/usecase"
)

func TestUpdate(t *testing.T) {
	t.Parallel()

	config := domain.TestConfig(t)
	ticket := domain.TestTicket(t)
	token := domain.TestToken(t)

	userRouter := router.New()
	// NOTE(toby3d): private resource
	userRouter.GET(ticket.Resource.URL().EscapedPath(), func(ctx *http.RequestCtx) {
		ctx.SuccessString(common.MIMETextHTMLCharsetUTF8,
			`<link rel="token_endpoint" href="https://auth.example.org/token">`,
		)
	})
	// NOTE(toby3d): token endpoint
	userRouter.POST("/token", func(ctx *http.RequestCtx) {
		ctx.SuccessString(common.MIMEApplicationJSONCharsetUTF8, `{
			"access_token": "`+token.AccessToken+`",
			"me": "`+token.Me.String()+`",
			"scope": "`+token.Scope.String()+`",
			"token_type": "Bearer"
		}`)
	})

	userClient, _, userCleanup := httptest.New(t, userRouter.Handler)
	t.Cleanup(userCleanup)

	router := router.New()
	delivery.NewRequestHandler(
		ucase.NewTicketUseCase(ticketrepo.NewMemoryTicketRepository(new(sync.Map), config), userClient, config),
		language.NewMatcher(message.DefaultCatalog.Languages()), config,
	).Register(router)

	client, _, cleanup := httptest.New(t, router.Handler)
	t.Cleanup(cleanup)

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
