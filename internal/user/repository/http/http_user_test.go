package http_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/fasthttp/router"
	"github.com/stretchr/testify/assert"
	http "github.com/valyala/fasthttp"

	"source.toby3d.me/toby3d/auth/internal/common"
	"source.toby3d.me/toby3d/auth/internal/domain"
	"source.toby3d.me/toby3d/auth/internal/testing/httptest"
	repository "source.toby3d.me/toby3d/auth/internal/user/repository/http"
)

const testBody string = `
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%[1]s</title>
  </head>
  <body>
    <div class="h-card">
      <img class="u-photo" src="%[3]s">
      <h1>
        <a class="p-name u-url" href="%[2]s">%[1]s</a>
      </h1>
      <a class="u-email" href="mailto:%[4]s">contact</a>
    </div>
  </body>
</html>
`

func TestGet(t *testing.T) {
	t.Parallel()

	user := domain.TestUser(t)
	client, _, cleanup := httptest.New(t, testHandler(t, user))
	t.Cleanup(cleanup)

	result, err := repository.NewHTTPUserRepository(client).Get(context.TODO(), user.Me)
	if err != nil {
		t.Fatal(err)
	}

	// NOTE(toby3d): endpoints
	assert.Equal(t, user.AuthorizationEndpoint.String(), result.AuthorizationEndpoint.String())
	assert.Equal(t, user.TokenEndpoint.String(), result.TokenEndpoint.String())
	assert.Equal(t, user.Micropub.String(), result.Micropub.String())
	assert.Equal(t, user.Microsub.String(), result.Microsub.String())

	// NOTE(toby3d): profile
	assert.Equal(t, user.Profile.Name, result.Profile.Name)
	assert.Equal(t, user.Profile.Email, result.Profile.Email)

	for i := range user.Profile.URL {
		assert.Equal(t, user.Profile.URL[i].String(), result.Profile.URL[i].String())
	}

	for i := range user.Profile.Photo {
		assert.Equal(t, user.Profile.Photo[i].String(), result.Profile.Photo[i].String())
	}
}

func testHandler(tb testing.TB, user *domain.User) http.RequestHandler {
	tb.Helper()

	router := router.New()
	router.GET("/", func(ctx *http.RequestCtx) {
		ctx.Response.Header.Set(http.HeaderLink, strings.Join([]string{
			`<` + user.AuthorizationEndpoint.String() + `>; rel="authorization_endpoint"`,
			`<` + user.IndieAuthMetadata.String() + `>; rel="indieauth-metadata"`,
			`<` + user.Micropub.String() + `>; rel="micropub"`,
			`<` + user.Microsub.String() + `>; rel="microsub"`,
			`<` + user.TicketEndpoint.String() + `>; rel="ticket_endpoint"`,
			`<` + user.TokenEndpoint.String() + `>; rel="token_endpoint"`,
		}, ", "))
		ctx.SuccessString(common.MIMETextHTMLCharsetUTF8, fmt.Sprintf(
			testBody, user.Name[0], user.URL[0].String(), user.Photo[0].String(), user.Email[0],
		))
	})
	router.GET(string(user.IndieAuthMetadata.Path()), func(ctx *http.RequestCtx) {
		ctx.SuccessString(common.MIMEApplicationJSONCharsetUTF8, `{
			"issuer": "`+user.Me.String()+`",
			"authorization_endpoint": "`+user.AuthorizationEndpoint.String()+`",
			"token_endpoint": "`+user.TokenEndpoint.String()+`"
		}`)
	})

	return router.Handler
}
