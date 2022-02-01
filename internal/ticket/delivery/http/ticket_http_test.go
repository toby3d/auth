package http_test

import (
	"sync"
	"testing"

	"github.com/fasthttp/router"
	http "github.com/valyala/fasthttp"
	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"source.toby3d.me/website/indieauth/internal/common"
	"source.toby3d.me/website/indieauth/internal/domain"
	"source.toby3d.me/website/indieauth/internal/testing/httptest"
	"source.toby3d.me/website/indieauth/internal/ticket"
	delivery "source.toby3d.me/website/indieauth/internal/ticket/delivery/http"
	ticketrepo "source.toby3d.me/website/indieauth/internal/ticket/repository/memory"
	ucase "source.toby3d.me/website/indieauth/internal/ticket/usecase"
)

type dependencies struct {
	client        *http.Client
	config        *domain.Config
	matcher       language.Matcher
	store         *sync.Map
	ticket        *domain.Ticket
	tickets       ticket.Repository
	ticketService ticket.UseCase
	token         *domain.Token
}

func TestUpdate(t *testing.T) {
	t.Parallel()

	deps := NewDependencies(t)

	r := router.New()
	delivery.NewRequestHandler(deps.ticketService, deps.matcher, deps.config).Register(r)

	client, _, cleanup := httptest.New(t, r.Handler)
	t.Cleanup(cleanup)

	const requestURI string = "https://example.com/ticket"

	req := httptest.NewRequest(http.MethodPost, requestURI, []byte(
		`ticket=`+deps.ticket.Ticket+
			`&resource=`+deps.ticket.Resource.String()+
			`&subject=`+deps.ticket.Subject.String(),
	))
	defer http.ReleaseRequest(req)
	req.Header.SetContentType(common.MIMEApplicationForm)
	domain.TestToken(t).SetAuthHeader(req)

	resp := http.AcquireResponse()
	defer http.ReleaseResponse(resp)

	if err := client.Do(req, resp); err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode() != http.StatusOK && resp.StatusCode() != http.StatusAccepted {
		t.Errorf("POST %s = %d, want %d or %d", requestURI, resp.StatusCode(), http.StatusOK,
			http.StatusAccepted)
	}

	// TODO(toby3d): print the result as part of the debugging. Instead, you
	// need to send or save the token to the recipient for later use.
	if resp.Body() == nil {
		t.Errorf("POST %s = nil, want something", requestURI)
	}
}

func NewDependencies(tb testing.TB) dependencies {
	tb.Helper()

	config := domain.TestConfig(tb)
	matcher := language.NewMatcher(message.DefaultCatalog.Languages())
	store := new(sync.Map)
	ticket := domain.TestTicket(tb)
	token := domain.TestToken(tb)

	r := router.New()
	// NOTE(toby3d): private resource
	r.GET(ticket.Resource.URL().EscapedPath(), func(ctx *http.RequestCtx) {
		ctx.SuccessString(common.MIMETextHTMLCharsetUTF8,
			`<link rel="token_endpoint" href="https://auth.example.org/token">`)
	})
	// NOTE(toby3d): token endpoint
	r.POST("/token", func(ctx *http.RequestCtx) {
		ctx.SuccessString(common.MIMEApplicationJSONCharsetUTF8, `{
			"access_token": "`+token.AccessToken+`",
			"me": "`+token.Me.String()+`",
			"scope": "`+token.Scope.String()+`",
			"token_type": "Bearer"
		}`)
	})

	client, _, cleanup := httptest.New(tb, r.Handler)
	tb.Cleanup(cleanup)

	tickets := ticketrepo.NewMemoryTicketRepository(store, config)
	ticketService := ucase.NewTicketUseCase(tickets, client, config)

	return dependencies{
		client:        client,
		config:        config,
		matcher:       matcher,
		store:         store,
		ticket:        ticket,
		tickets:       tickets,
		ticketService: ticketService,
		token:         token,
	}
}
