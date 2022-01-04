package http

import (
	"encoding/json"

	"github.com/fasthttp/router"
	http "github.com/valyala/fasthttp"

	"source.toby3d.me/website/indieauth/internal/common"
	"source.toby3d.me/website/indieauth/internal/domain"
)

type (
	//nolint: tagliatelle
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
		Issuer string `json:"issuer"`

		// The Authorization Endpoint.
		AuthorizationEndpoint string `json:"authorization_endpoint"`

		// The Token Endpoint.
		TokenEndpoint string `json:"token_endpoint"`

		//  JSON array containing scope values supported by the
		// IndieAuth server. Servers MAY choose not to advertise some
		// supported scope values even when this parameter is used.
		ScopesSupported []string `json:"scopes_supported,omitempty"`

		// JSON array containing the response_type values supported.
		// This differs from RFC8414 in that this parameter is OPTIONAL
		// and that, if omitted, the default is code.
		ResponseTypesSupported []string `json:"response_types_supported,omitempty"`

		// JSON array containing grant type values supported. If
		// omitted, the default value differs from RFC8414 and is
		// authorization_code.
		GrantTypesSupported []string `json:"grant_types_supported,omitempty"`

		// URL of a page containing human-readable information that
		// developers might need to know when using the server. This
		// might be a link to the IndieAuth spec or something more
		// personal to your implementation.
		ServiceDocumentation string `json:"service_documentation,omitempty"`

		// JSON array containing the methods supported for PKCE. This
		// parameter differs from RFC8414 in that it is not optional as
		// PKCE is REQUIRED.
		CodeChallengeMethodsSupported []string `json:"code_challenge_methods_supported"`

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
		config *domain.Config
	}
)

//nolint: gochecknoglobals // NOTE(toby3d): structs cannot be contants.
var DefaultMetadataResponse = MetadataResponse{
	ServiceDocumentation:                       "https://indieauth.spec.indieweb.org/",
	AuthorizationResponseIssParameterSupported: true,
	ScopesSupported: []string{
		domain.ScopeEmail.String(),
		domain.ScopeProfile.String(),
	},
	CodeChallengeMethodsSupported: []string{
		domain.CodeChallengeMethodMD5.String(),
		domain.CodeChallengeMethodPLAIN.String(),
		domain.CodeChallengeMethodS1.String(),
		domain.CodeChallengeMethodS256.String(),
		domain.CodeChallengeMethodS512.String(),
	},
	ResponseTypesSupported: []string{
		domain.ResponseTypeCode.String(),
		domain.ResponseTypeID.String(),
	},
	GrantTypesSupported: []string{
		domain.GrantTypeAuthorizationCode.String(),
	},
}

func NewRequestHandler(config *domain.Config) *RequestHandler {
	return &RequestHandler{
		config: config,
	}
}

func (h *RequestHandler) Register(r *router.Router) {
	r.GET("/.well-known/oauth-authorization-server", h.read)
}

func (h *RequestHandler) read(ctx *http.RequestCtx) {
	resp := DefaultMetadataResponse
	resp.Issuer = h.config.Server.GetRootURL()
	resp.AuthorizationEndpoint = resp.Issuer + "authorize"
	resp.TokenEndpoint = resp.Issuer + "token"

	ctx.SetStatusCode(http.StatusOK)
	ctx.SetContentType(common.MIMEApplicationJSON)
	json.NewEncoder(ctx).Encode(&resp)
}
