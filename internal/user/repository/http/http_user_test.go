package http_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"source.toby3d.me/toby3d/auth/internal/common"
	"source.toby3d.me/toby3d/auth/internal/domain"
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

	srv := httptest.NewServer(testHandler(t, user))
	t.Cleanup(srv.Close)

	user.Me = domain.TestMe(t, srv.URL+"/")

	result, err := repository.NewHTTPUserRepository(srv.Client()).
		Get(context.Background(), *user.Me)
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(user, result, cmp.AllowUnexported(domain.Me{}, domain.Email{})); diff != "" {
		t.Errorf("%+s", diff)
	}
}

func testHandler(tb testing.TB, user *domain.User) http.Handler {
	tb.Helper()

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(common.HeaderLink, strings.Join([]string{
			`<` + user.AuthorizationEndpoint.String() + `>; rel="authorization_endpoint"`,
			`<` + user.IndieAuthMetadata.String() + `>; rel="indieauth-metadata"`,
			`<` + user.Micropub.String() + `>; rel="micropub"`,
			`<` + user.Microsub.String() + `>; rel="microsub"`,
			`<` + user.TicketEndpoint.String() + `>; rel="ticket_endpoint"`,
			`<` + user.TokenEndpoint.String() + `>; rel="token_endpoint"`,
		}, ", "))
		w.Header().Set(common.HeaderContentType, common.MIMETextHTMLCharsetUTF8)
		fmt.Fprintf(w, testBody, user.Name[0], user.URL[0].String(), user.Photo[0].String(), user.Email[0])
	})
	mux.HandleFunc(user.IndieAuthMetadata.Path, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(common.HeaderContentType, common.MIMEApplicationJSONCharsetUTF8)
		fmt.Fprint(w, `{
			"issuer": "`+user.Me.String()+`",
			"authorization_endpoint": "`+user.AuthorizationEndpoint.String()+`",
			"token_endpoint": "`+user.TokenEndpoint.String()+`"
		}`)
	})

	return mux
}
