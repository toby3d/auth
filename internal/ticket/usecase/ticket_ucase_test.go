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
	repo "source.toby3d.me/website/oauth/internal/ticket/repository/memory"
	ucase "source.toby3d.me/website/oauth/internal/ticket/usecase"
)

func TestRedeem(t *testing.T) {
	t.Parallel()

	token := domain.TestToken(t)
	ticket := domain.TestTicket(t)

	store := new(sync.Map)
	store.Store(
		path.Join(repo.DefaultPathPrefix, ticket.Resource.String()),
		domain.TestURL(t, "https://example.com/token"),
	)

	client, _, cleanup := httptest.New(t, func(ctx *http.RequestCtx) {
		ctx.SuccessString(common.MIMEApplicationJSONCharsetUTF8, fmt.Sprintf(`{
			"token_type": "Bearer",
			"access_token": "%s",
			"scope": "%s",
			"me": "%s"
		}`, token.AccessToken, token.Scope.String(), token.Me.String()))
	})
	t.Cleanup(cleanup)

	result, err := ucase.NewTicketUseCase(repo.NewMemoryTicketRepository(store), client).
		Redeem(context.Background(), ticket)
	require.NoError(t, err)
	assert.Equal(t, token.AccessToken, result.AccessToken)
	assert.Equal(t, token.Me.String(), result.Me.String())
	assert.Equal(t, token.Scope, result.Scope)
}
