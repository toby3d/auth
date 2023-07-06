package http_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"

	repository "source.toby3d.me/toby3d/auth/internal/client/repository/http"
	"source.toby3d.me/toby3d/auth/internal/common"
	"source.toby3d.me/toby3d/auth/internal/domain"
)

const testBody string = `<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%[1]s</title>
    <link rel="redirect_uri" href="%[4]s">
  </head>
  <body>
    <div class="h-app h-x-app">
      <img class="u-logo" src="%[3]s">
      <a class="u-url p-name" href="%[2]s">%[1]s</a>
    </div>
  </body>
</html>
`

func TestGet(t *testing.T) {
	t.Parallel()

	client := domain.TestClient(t)
	srv := httptest.NewUnstartedServer(testHandler(t, *client))
	srv.EnableHTTP2 = true

	srv.StartTLS()
	t.Cleanup(srv.Close)

	client.ID = *domain.TestClientID(t, srv.URL+"/")
	clients := repository.NewHTTPClientRepository(srv.Client())

	result, err := clients.Get(context.Background(), client.ID)
	if err != nil {
		t.Fatal(err)
	}

	if out := client.ID; !result.ID.IsEqual(out) {
		t.Errorf("GET %s = %s, want %s", client.ID, out, result.ID)
	}

	if !cmp.Equal(result.Name, client.Name) {
		t.Errorf("GET %s = %+s, want %+s", client.ID, result.Name, client.Name)
	}

	if !cmp.Equal(result.URL, client.URL) {
		t.Errorf("GET %s = %+s, want %+s", client.ID, result.URL, client.URL)
	}

	if !cmp.Equal(result.Logo, client.Logo) {
		t.Errorf("GET %s = %+s, want %+s", client.ID, result.Logo, client.Logo)
	}

	if !cmp.Equal(result.RedirectURI, client.RedirectURI) {
		t.Errorf("GET %s = %+s, want %+s", client.ID, result.RedirectURI, client.RedirectURI)
	}
}

func testHandler(tb testing.TB, client domain.Client) http.Handler {
	tb.Helper()

	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set(common.HeaderContentType, common.MIMETextHTMLCharsetUTF8)
		w.Header().Set(common.HeaderLink, `<`+client.RedirectURI[0].String()+`>; rel="redirect_uri"`)
		fmt.Fprintf(w, testBody, client.Name[0], client.URL[0], client.Logo[0], client.RedirectURI[1])
	})
}
