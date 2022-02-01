package usecase_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/fasthttp/router"
	http "github.com/valyala/fasthttp"

	"source.toby3d.me/website/indieauth/internal/common"
	"source.toby3d.me/website/indieauth/internal/domain"
	"source.toby3d.me/website/indieauth/internal/testing/httptest"
	ucase "source.toby3d.me/website/indieauth/internal/ticket/usecase"
)

func TestRedeem(t *testing.T) {
	t.Parallel()

	token := domain.TestToken(t)
	ticket := domain.TestTicket(t)

	router := router.New()
	router.GET(string(ticket.Resource.Path()), func(ctx *http.RequestCtx) {
		ctx.SuccessString(common.MIMETextHTMLCharsetUTF8, `<link rel="token_endpoint" href="`+
			ticket.Subject.String()+`token">`)
	})
	router.POST("/token", func(ctx *http.RequestCtx) {
		ctx.SuccessString(common.MIMEApplicationJSONCharsetUTF8, fmt.Sprintf(`{
			"token_type": "Bearer",
			"access_token": "%s",
			"scope": "%s",
			"me": "%s"
		}`, token.AccessToken, token.Scope.String(), token.Me.String()))
	})

	client, _, cleanup := httptest.New(t, router.Handler)
	t.Cleanup(cleanup)

	result, err := ucase.NewTicketUseCase(nil, client, domain.TestConfig(t)).
		Redeem(context.TODO(), ticket)
	if err != nil {
		t.Fatal(err)
	}

	if result.String() != token.String() {
		t.Errorf("Redeem(%+v) = %s, want %s", ticket, result, token)
	}
}
