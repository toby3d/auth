package http

import (
	"github.com/fasthttp/router"
	"github.com/goccy/go-json"
	http "github.com/valyala/fasthttp"

	"source.toby3d.me/toby3d/middleware"
	"source.toby3d.me/website/indieauth/internal/common"
	"source.toby3d.me/website/indieauth/internal/domain"
)

type (
	//nolint: tagliatelle // https://indieauth.net/source/#indieauth-server-metadata
	MetadataResponse struct {
		// The server's issuer identifier. The issuer identifier is a
		// URL that uses the "https" scheme and has no query or fragment
		// components. The identifier MUST be a prefix of the
		// indieauth-metadata URL. e.g. for an indieauth-metadata
		// endpoint
		// https://example.com/.well-known/oauth-authorization-server,
		// the issuer URL could be https://example.com/, or for a
		// metadata URL of
		// https://example.com/wp-json/indieauth/1.0/metadata, the
		// issuer URL could be https://example.com/wp-json/indieauth/1.0
		Issuer *domain.ClientID `json:"issuer"`

		// The Authorization Endpoint.
		AuthorizationEndpoint *domain.URL `json:"authorization_endpoint"`

		// The Token Endpoint.
		TokenEndpoint *domain.URL `json:"token_endpoint"`

		// JSON array containing scope values supported by the
		// IndieAuth server. Servers MAY choose not to advertise some
		// supported scope values even when this parameter is used.
		ScopesSupported []domain.Scope `json:"scopes_supported,omitempty"`

		// JSON array containing the response_type values supported.
		// This differs from RFC8414 in that this parameter is OPTIONAL
		// and that, if omitted, the default is code.
		ResponseTypesSupported []domain.ResponseType `json:"response_types_supported,omitempty"`

		// JSON array containing grant type values supported. If
		// omitted, the default value differs from RFC8414 and is
		// authorization_code.
		GrantTypesSupported []domain.GrantType `json:"grant_types_supported,omitempty"`

		// URL of a page containing human-readable information that
		// developers might need to know when using the server. This
		// might be a link to the IndieAuth spec or something more
		// personal to your implementation.
		ServiceDocumentation *domain.URL `json:"service_documentation,omitempty"`

		// JSON array containing the methods supported for PKCE. This
		// parameter differs from RFC8414 in that it is not optional as
		// PKCE is REQUIRED.
		CodeChallengeMethodsSupported []domain.CodeChallengeMethod `json:"code_challenge_methods_supported"`

		// Boolean parameter indicating whether the authorization server
		// provides the iss parameter. If omitted, the default value is
		// false. As the iss parameter is REQUIRED, this is provided for
		// compatibility with OAuth 2.0 servers implementing the
		// parameter.
		//
		//nolint: lll
		AuthorizationResponseIssParameterSupported bool `json:"authorization_response_iss_parameter_supported,omitempty"`
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
	_ = json.NewEncoder(ctx).Encode(&MetadataResponse{
		Issuer:                        h.metadata.Issuer,
		AuthorizationEndpoint:         h.metadata.AuthorizationEndpoint,
		TokenEndpoint:                 h.metadata.TokenEndpoint,
		ScopesSupported:               h.metadata.ScopesSupported,
		ResponseTypesSupported:        h.metadata.ResponseTypesSupported,
		GrantTypesSupported:           h.metadata.GrantTypesSupported,
		ServiceDocumentation:          h.metadata.ServiceDocumentation,
		CodeChallengeMethodsSupported: h.metadata.CodeChallengeMethodsSupported,
		AuthorizationResponseIssParameterSupported: h.metadata.AuthorizationResponseIssParameterSupported,
	})
}
