package http_test

import (
	"reflect"
	"testing"

	"github.com/fasthttp/router"
	"github.com/goccy/go-json"
	http "github.com/valyala/fasthttp"

	"source.toby3d.me/website/indieauth/internal/domain"
	delivery "source.toby3d.me/website/indieauth/internal/metadata/delivery/http"
	"source.toby3d.me/website/indieauth/internal/testing/httptest"
)

func TestMetadata(t *testing.T) {
	t.Parallel()

	r := router.New()
	cfg := domain.TestConfig(t)
	delivery.NewRequestHandler(cfg).Register(r)

	client, _, cleanup := httptest.New(t, r.Handler)
	t.Cleanup(cleanup)

	const requestURL string = "https://example.com/.well-known/oauth-authorization-server"

	status, body, err := client.Get(nil, requestURL)
	if err != nil {
		t.Fatal(err)
	}

	if status != http.StatusOK {
		t.Errorf("GET %s = %d, want %d", requestURL, status, http.StatusOK)
	}

	result := new(delivery.MetadataResponse)
	if err = json.Unmarshal(body, result); err != nil {
		t.Fatal(err)
	}

	expResult := delivery.DefaultMetadataResponse
	expResult.Issuer = cfg.Server.GetRootURL()
	expResult.AuthorizationEndpoint = expResult.Issuer + "authorize"
	expResult.TokenEndpoint = expResult.Issuer + "token"

	if !reflect.DeepEqual(*result, expResult) {
		t.Errorf("Unmarshal(%s) = %+v, want %+v", body, result, expResult)
	}
}
