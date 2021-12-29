package usecase_test

import (
	"context"
	"fmt"
	"path"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	http "github.com/valyala/fasthttp"

	"source.toby3d.me/website/oauth/internal/common"
	"source.toby3d.me/website/oauth/internal/domain"
	"source.toby3d.me/website/oauth/internal/testing/httptest"
	ucase "source.toby3d.me/website/oauth/internal/ticket/usecase"
	userrepo "source.toby3d.me/website/oauth/internal/user/repository/memory"
)

func TestRedeem(t *testing.T) {
	t.Parallel()

	token := domain.TestToken(t)
	ticket := domain.TestTicket(t)

	store := new(sync.Map)
	store.Store(path.Join(userrepo.DefaultPathPrefix, ticket.Subject.String()), domain.TestUser(t))

	client, _, cleanup := httptest.New(t, func(ctx *http.RequestCtx) {
		ctx.SuccessString(common.MIMEApplicationJSONCharsetUTF8, fmt.Sprintf(`{
			"access_token": "%s",
			"token_type": "Bearer",
			"scope": "%s",
			"me": "%s"
		}`, token.AccessToken, token.Scope.String(), token.Me.String()))
	})
	t.Cleanup(cleanup)

	result, err := ucase.NewTicketUseCase(client).
		Redeem(context.Background(), domain.TestURL(t, "https://bob.example.com/token"), ticket.Ticket)
	require.NoError(t, err)
	assert.Equal(t, token.AccessToken, result.AccessToken)
	assert.Equal(t, token.Me.String(), result.Me.String())
	assert.Equal(t, token.Scope, result.Scope)
}
