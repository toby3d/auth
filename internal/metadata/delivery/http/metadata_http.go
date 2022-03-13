package http

import (
	"github.com/fasthttp/router"
	"github.com/goccy/go-json"
	http "github.com/valyala/fasthttp"

	"source.toby3d.me/toby3d/middleware"
	"source.toby3d.me/toby3d/auth/internal/common"
	"source.toby3d.me/toby3d/auth/internal/domain"
)

type (
	//nolint: tagliatelle // https://indieauth.net/source/#indieauth-server-metadata
	MetadataResponse struct {
		// The server's issuer identifier.
		Issuer string `json:"issuer"`

		// The Authorization Endpoint.
		AuthorizationEndpoint string `json:"authorization_endpoint"`

		// The Token Endpoint.
		TokenEndpoint string `json:"token_endpoint"`

		// The Introspection Endpoint.
		IntrospectionEndpoint string `json:"introspection_endpoint"`

		// JSON array containing a list of client authentication methods
		// supported by this introspection endpoint.
		IntrospectionEndpointAuthMethodsSupported []string `json:"introspection_endpoint_auth_methods_supported,omitempty"` //nolint: lll

		// The Revocation Endpoint.
		RevocationEndpoint string `json:"revocation_endpoint,omitempty"`

		// JSON array containing the value "none".
		RevocationEndpointAuthMethodsSupported []string `json:"revocation_endpoint_auth_methods_supported,omitempty"` //nolint: lll

		// JSON array containing scope values supported by the
		// IndieAuth server.
		ScopesSupported []string `json:"scopes_supported,omitempty"`

		// JSON array containing the response_type values supported.
		ResponseTypesSupported []string `json:"response_types_supported,omitempty"`

		// JSON array containing grant type values supported.
		GrantTypesSupported []string `json:"grant_types_supported,omitempty"`

		// URL of a page containing human-readable information that
		// developers might need to know when using the server.
		ServiceDocumentation string `json:"service_documentation,omitempty"`

		// JSON array containing the methods supported for PKCE.
		CodeChallengeMethodsSupported []string `json:"code_challenge_methods_supported"`

		// Boolean parameter indicating whether the authorization server
		// provides the iss parameter.
		AuthorizationResponseIssParameterSupported bool `json:"authorization_response_iss_parameter_supported,omitempty"` //nolint: lll

		// The User Info Endpoint.
		UserinfoEndpoint string `json:"userinfo_endpoint,omitempty"`
	}

	RequestHandler struct {
		metadata *domain.Metadata
	}
)

func NewRequestHandler(metadata *domain.Metadata) *RequestHandler {
	return &RequestHandler{
		metadata: metadata,
	}
}

func (h *RequestHandler) Register(r *router.Router) {
	chain := middleware.Chain{
		middleware.LogFmt(),
	}

	r.GET("/.well-known/oauth-authorization-server", chain.RequestHandler(h.read))
}

func (h *RequestHandler) read(ctx *http.RequestCtx) {
	ctx.SetStatusCode(http.StatusOK)
	ctx.SetContentType(common.MIMEApplicationJSONCharsetUTF8)

	scopes, responseTypes, grantTypes, codeChallengeMethods := make([]string, 0), make([]string, 0),
		make([]string, 0), make([]string, 0)

	for i := range h.metadata.ScopesSupported {
		scopes = append(scopes, h.metadata.ScopesSupported[i].String())
	}

	for i := range h.metadata.ResponseTypesSupported {
		responseTypes = append(responseTypes, h.metadata.ResponseTypesSupported[i].String())
	}

	for i := range h.metadata.GrantTypesSupported {
		grantTypes = append(grantTypes, h.metadata.GrantTypesSupported[i].String())
	}

	for i := range h.metadata.CodeChallengeMethodsSupported {
		codeChallengeMethods = append(codeChallengeMethods,
			h.metadata.CodeChallengeMethodsSupported[i].String())
	}

	_ = json.NewEncoder(ctx).Encode(&MetadataResponse{
		AuthorizationEndpoint: h.metadata.AuthorizationEndpoint.String(),
		IntrospectionEndpoint: h.metadata.IntrospectionEndpoint.String(),
		Issuer:                h.metadata.Issuer.String(),
		RevocationEndpoint:    h.metadata.RevocationEndpoint.String(),
		ServiceDocumentation:  h.metadata.ServiceDocumentation.String(),
		TokenEndpoint:         h.metadata.TokenEndpoint.String(),
		UserinfoEndpoint:      h.metadata.UserinfoEndpoint.String(),
		AuthorizationResponseIssParameterSupported: h.metadata.AuthorizationResponseIssParameterSupported,
		CodeChallengeMethodsSupported:              codeChallengeMethods,
		GrantTypesSupported:                        grantTypes,
		IntrospectionEndpointAuthMethodsSupported:  h.metadata.IntrospectionEndpointAuthMethodsSupported,
		ResponseTypesSupported:                     responseTypes,
		ScopesSupported:                            scopes,
		// NOTE(toby3d): If a revocation endpoint is provided, this
		// property should also be provided with the value ["none"],
		// since the omission of this value defaults to
		// client_secret_basic according to RFC8414.
		RevocationEndpointAuthMethodsSupported: h.metadata.RevocationEndpointAuthMethodsSupported,
	})
}
