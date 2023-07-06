package http_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/goccy/go-json"
	"github.com/google/go-cmp/cmp"

	"source.toby3d.me/toby3d/auth/internal/common"
	"source.toby3d.me/toby3d/auth/internal/domain"
	repository "source.toby3d.me/toby3d/auth/internal/metadata/repository/http"
)

//nolint:lll,tagliatelle
type Response struct {
	UserinfoEndpoint                           string   `json:"userinfo_endpoint,omitempty"`
	AuthorizationEndpoint                      string   `json:"authorization_endpoint"`
	IntrospectionEndpoint                      string   `json:"introspection_endpoint"`
	Microsub                                   string   `json:"microsub"`
	RevocationEndpoint                         string   `json:"revocation_endpoint,omitempty"`
	Micropub                                   string   `json:"micropub"`
	Issuer                                     string   `json:"issuer"`
	ServiceDocumentation                       string   `json:"service_documentation,omitempty"`
	TicketEndpoint                             string   `json:"ticket_endpoint"`
	TokenEndpoint                              string   `json:"token_endpoint"`
	RevocationEndpointAuthMethodsSupported     []string `json:"revocation_endpoint_auth_methods_supported,omitempty"`
	IntrospectionEndpointAuthMethodsSupported  []string `json:"introspection_endpoint_auth_methods_supported,omitempty"`
	CodeChallengeMethodsSupported              []string `json:"code_challenge_methods_supported"`
	ResponseTypesSupported                     []string `json:"response_types_supported,omitempty"`
	GrantTypesSupported                        []string `json:"grant_types_supported,omitempty"`
	ScopesSupported                            []string `json:"scopes_supported,omitempty"`
	AuthorizationResponseIssParameterSupported bool     `json:"authorization_response_iss_parameter_supported,omitempty"`
}

const testBody string = `<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="utf-8">
    <title>Testing</title>
    %s
  </head>
  <body></body>
</html>`

//nolint:funlen
func TestGet(t *testing.T) {
	t.Parallel()

	testMetadata := domain.TestMetadata(t)

	for _, tc := range []struct {
		header map[string]string
		body   map[string]string
		out    *domain.Metadata
		name   string
	}{
		{
			name: "header",
			header: map[string]string{
				"indieauth-metadata":     "/metadata",
				"authorization_endpoint": "http://example.net/authorization",
				"token_endpoint":         "http://example.net/tkn",
			},
			out: testMetadata,
		}, /*{
			name: "body",
			body: map[string]string{
				"indieauth-metadata":     "/metadata",
				"authorization_endpoint": "http://example.net/authorization",
				"token_endpoint":         "http://example.net/tkn",
			},
			out: &testMetadata,
		}*/} {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mux := http.NewServeMux()
			mux.HandleFunc("/metadata", func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set(common.HeaderContentType, common.MIMEApplicationJSONCharsetUTF8)
				_ = json.NewEncoder(w).Encode(NewResponse(t, *testMetadata))
			})
			mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
				links := make([]string, 0)
				for k, v := range tc.header {
					links = append(links, `<`+v+`>; rel="`+k+`"`)
				}

				w.Header().Set(common.HeaderLink, strings.Join(links, ", "))

				links = make([]string, 0)
				for k, v := range tc.body {
					links = append(links, `<link rel="`+k+`" href="`+v+`">`)
				}

				fmt.Fprintf(w, testBody, strings.Join(links, "\n"))
			})

			srv := httptest.NewUnstartedServer(mux)
			srv.EnableHTTP2 = true
			srv.Start()
			t.Cleanup(srv.Close)

			tc.header["indieauth-metadata"] = srv.URL + tc.header["indieauth-metadata"]

			u, _ := url.Parse(srv.URL + "/")
			out, err := repository.NewHTTPMetadataRepository(srv.Client()).
				Get(context.Background(), u)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(tc.out, out, cmp.AllowUnexported(
				domain.ClientID{},
				domain.CodeChallengeMethod{},
				domain.GrantType{},
				domain.ResponseType{},
				domain.Scope{},
				url.URL{},
			)); diff != "" {
				t.Errorf("%+s", diff)
			}
		})
	}
}

func NewResponse(tb testing.TB, src domain.Metadata) *Response {
	tb.Helper()

	out := &Response{
		CodeChallengeMethodsSupported:              make([]string, 0),
		GrantTypesSupported:                        make([]string, 0),
		ResponseTypesSupported:                     make([]string, 0),
		ScopesSupported:                            make([]string, 0),
		IntrospectionEndpointAuthMethodsSupported:  make([]string, 0),
		RevocationEndpointAuthMethodsSupported:     make([]string, 0),
		Issuer:                                     src.Issuer.String(),
		AuthorizationEndpoint:                      src.AuthorizationEndpoint.String(),
		IntrospectionEndpoint:                      src.IntrospectionEndpoint.String(),
		RevocationEndpoint:                         src.RevocationEndpoint.String(),
		ServiceDocumentation:                       src.ServiceDocumentation.String(),
		TokenEndpoint:                              src.TokenEndpoint.String(),
		UserinfoEndpoint:                           src.UserinfoEndpoint.String(),
		TicketEndpoint:                             src.TicketEndpoint.String(),
		Micropub:                                   src.MicropubEndpoint.String(),
		Microsub:                                   src.MicrosubEndpoint.String(),
		AuthorizationResponseIssParameterSupported: src.AuthorizationResponseIssParameterSupported,
	}

	for _, method := range src.CodeChallengeMethodsSupported {
		out.CodeChallengeMethodsSupported = append(out.CodeChallengeMethodsSupported, method.String())
	}

	for _, grantType := range src.GrantTypesSupported {
		out.GrantTypesSupported = append(out.GrantTypesSupported, grantType.String())
	}

	for _, responseType := range src.ResponseTypesSupported {
		out.ResponseTypesSupported = append(out.ResponseTypesSupported, responseType.String())
	}

	for _, scope := range src.ScopesSupported {
		out.ScopesSupported = append(out.ScopesSupported, scope.String())
	}

	out.IntrospectionEndpointAuthMethodsSupported = append(out.IntrospectionEndpointAuthMethodsSupported,
		src.IntrospectionEndpointAuthMethodsSupported...)

	out.RevocationEndpointAuthMethodsSupported = append(out.RevocationEndpointAuthMethodsSupported,
		src.RevocationEndpointAuthMethodsSupported...)

	return out
}
