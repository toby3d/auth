package http_test

/* TODO(toby3d): move CSRF middleware into main
import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"source.toby3d.me/toby3d/auth/internal/common"
	"source.toby3d.me/toby3d/auth/internal/domain"
	"source.toby3d.me/toby3d/auth/internal/ticket"
	delivery "source.toby3d.me/toby3d/auth/internal/ticket/delivery/http"
	ticketrepo "source.toby3d.me/toby3d/auth/internal/ticket/repository/memory"
	ucase "source.toby3d.me/toby3d/auth/internal/ticket/usecase"
)

type Dependencies struct {
	server        *httptest.Server
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
	t.Cleanup(deps.server.Close)

	req := httptest.NewRequest(http.MethodPost, "https://example.com/", strings.NewReader(
		`ticket=`+deps.ticket.Ticket+
			`&resource=`+deps.ticket.Resource.String()+
			`&subject=`+deps.ticket.Subject.String(),
	))
	req.Header.Set(common.HeaderContentType, common.MIMEApplicationForm)
	deps.token.SetAuthHeader(req)

	w := httptest.NewRecorder()
	delivery.NewHandler(deps.ticketService, deps.matcher, *deps.config).
		Handler().
		ServeHTTP(w, req)

	domain.TestToken(t).SetAuthHeader(req)

	resp := w.Result()

	if resp.StatusCode != http.StatusOK &&
		resp.StatusCode != http.StatusAccepted {
		t.Errorf("%s %s = %d, want %d or %d", req.Method, req.RequestURI, resp.StatusCode, http.StatusOK,
			http.StatusAccepted)
	}

	// TODO(toby3d): print the result as part of the debugging. Instead, you
	// need to send or save the token to the recipient for later use.
	if resp.Body == nil {
		t.Errorf("%s %s = nil, want not nil", req.Method, req.RequestURI)
	}
}

func NewDependencies(tb testing.TB) Dependencies {
	tb.Helper()

	config := domain.TestConfig(tb)
	matcher := language.NewMatcher(message.DefaultCatalog.Languages())
	store := new(sync.Map)
	ticket := domain.TestTicket(tb)
	token := domain.TestToken(tb)

	mux := http.NewServeMux()
	// NOTE(toby3d): private resource
	mux.HandleFunc(ticket.Resource.Path, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(common.HeaderContentType, common.MIMETextHTMLCharsetUTF8)
		fmt.Fprintf(w, `<link rel="token_endpoint" href="https://auth.example.org/token">`)
	})
	// NOTE(toby3d): token endpoint
	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)

			return
		}

		w.Header().Set(common.HeaderContentType, common.MIMEApplicationJSONCharsetUTF8)
		fmt.Fprintf(w, `{
				"access_token": "`+token.AccessToken+`",
				"me": "`+token.Me.String()+`",
				"scope": "`+token.Scope.String()+`",
				"token_type": "Bearer"
			}`)
	})

	server := httptest.NewServer(mux)
	client := server.Client()
	tickets := ticketrepo.NewMemoryTicketRepository(store, config)
	ticketService := ucase.NewTicketUseCase(tickets, client, config)

	return Dependencies{
		server:        server,
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
*/
