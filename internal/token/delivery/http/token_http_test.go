package http_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/goccy/go-json"

	"source.toby3d.me/toby3d/auth/internal/common"
	"source.toby3d.me/toby3d/auth/internal/domain"
	"source.toby3d.me/toby3d/auth/internal/profile"
	profilerepo "source.toby3d.me/toby3d/auth/internal/profile/repository/memory"
	"source.toby3d.me/toby3d/auth/internal/session"
	sessionrepo "source.toby3d.me/toby3d/auth/internal/session/repository/memory"
	"source.toby3d.me/toby3d/auth/internal/ticket"
	ticketrepo "source.toby3d.me/toby3d/auth/internal/ticket/repository/memory"
	ticketucase "source.toby3d.me/toby3d/auth/internal/ticket/usecase"
	"source.toby3d.me/toby3d/auth/internal/token"
	delivery "source.toby3d.me/toby3d/auth/internal/token/delivery/http"
	tokenrepo "source.toby3d.me/toby3d/auth/internal/token/repository/memory"
	tokenucase "source.toby3d.me/toby3d/auth/internal/token/usecase"
)

type Dependencies struct {
	client        *http.Client
	config        *domain.Config
	profiles      profile.Repository
	sessions      session.Repository
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

func TestIntrospection(t *testing.T) {
	t.Parallel()

	deps := NewDependencies(t)

	req := httptest.NewRequest(http.MethodPost, "https://app.example.com/introspect",
		strings.NewReader("token="+deps.token.AccessToken))
	req.Header.Set(common.HeaderAccept, common.MIMEApplicationJSON)
	req.Header.Set(common.HeaderContentType, common.MIMEApplicationForm)

	w := httptest.NewRecorder()
	delivery.NewHandler(deps.tokenService, deps.ticketService, deps.config).
		Handler().
		ServeHTTP(w, req)

	resp := w.Result()

	if result := resp.StatusCode; result != http.StatusOK {
		t.Errorf("%s %s = %d, want %d", req.Method, req.RequestURI, result, http.StatusOK)
	}

	result := new(delivery.TokenIntrospectResponse)
	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		t.Fatal(err)
	}

	deps.token.AccessToken = ""

	if result.ClientID != deps.token.ClientID.String() ||
		result.Me != deps.token.Me.String() ||
		result.Scope != deps.token.Scope.String() {
		t.Errorf("%s %s = %+v, want %+v", req.Method, req.RequestURI, result, deps.token)
	}
}

func TestRevocation(t *testing.T) {
	t.Parallel()

	deps := NewDependencies(t)

	req := httptest.NewRequest(http.MethodPost, "https://app.example.com/revocation",
		strings.NewReader(`token=`+deps.token.AccessToken))
	req.Header.Set(common.HeaderContentType, common.MIMEApplicationForm)
	req.Header.Set(common.HeaderAccept, common.MIMEApplicationJSON)

	w := httptest.NewRecorder()
	delivery.NewHandler(deps.tokenService, deps.ticketService, deps.config).
		Handler().
		ServeHTTP(w, req)

	resp := w.Result()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("%s %s = %d, want %d", req.Method, req.RequestURI, resp.StatusCode, http.StatusOK)
	}

	expBody := []byte("{}") //nolint:ifshort
	if result := bytes.TrimSpace(body); !bytes.Equal(result, expBody) {
		t.Errorf("%s %s = %s, want %s", req.Method, req.RequestURI, result, expBody)
	}

	result, err := deps.tokens.Get(context.Background(), deps.token.AccessToken)
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
	token := domain.TestToken(tb)
	profiles := profilerepo.NewMemoryProfileRepository()
	sessions := sessionrepo.NewMemorySessionRepository(*config)
	tickets := ticketrepo.NewMemoryTicketRepository(*config)
	tokens := tokenrepo.NewMemoryTokenRepository()
	ticketService := ticketucase.NewTicketUseCase(tickets, client, config)
	tokenService := tokenucase.NewTokenUseCase(tokenucase.Config{
		Config:   config,
		Profiles: profiles,
		Sessions: sessions,
		Tokens:   tokens,
	})

	return Dependencies{
		client:        client,
		config:        config,
		profiles:      profiles,
		sessions:      sessions,
		tickets:       tickets,
		ticketService: ticketService,
		token:         token,
		tokens:        tokens,
		tokenService:  tokenService,
	}
}
