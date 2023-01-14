package usecase_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"source.toby3d.me/toby3d/auth/internal/common"
	"source.toby3d.me/toby3d/auth/internal/domain"
	ucase "source.toby3d.me/toby3d/auth/internal/ticket/usecase"
)

func TestRedeem(t *testing.T) {
	t.Parallel()

	token := domain.TestToken(t)
	ticket := domain.TestTicket(t)

	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)

			return
		}

		w.Header().Set(common.HeaderContentType, common.MIMEApplicationJSONCharsetUTF8)
		fmt.Fprintf(w, `{
			"token_type": "Bearer",
			"access_token": "%s",
			"scope": "%s",
			"me": "%s"
		}`, token.AccessToken, token.Scope.String(), token.Me.String())
	}))
	t.Cleanup(tokenServer.Close)

	subjectServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(common.HeaderContentType, common.MIMETextHTMLCharsetUTF8)
		fmt.Fprint(w, `<link rel="token_endpoint" href="`+tokenServer.URL+`/token">`)
	}))
	t.Cleanup(subjectServer.Close)

	ticket.Resource, _ = url.Parse(subjectServer.URL + "/")

	result, err := ucase.NewTicketUseCase(nil, subjectServer.Client(), domain.TestConfig(t)).
		Redeem(context.Background(), *ticket)
	if err != nil {
		t.Fatal(err)
	}

	if result.String() != token.String() {
		t.Errorf("Redeem(%+v) = %s, want %s", ticket, result, token)
	}
}
