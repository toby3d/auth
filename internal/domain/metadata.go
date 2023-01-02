package domain

import (
	"net/url"
	"testing"
)

type Metadata struct {
	// The server's issuer identifier. The issuer identifier is a URL that
	// uses the "https" scheme and has no query or fragment components. The
	// identifier MUST be a prefix of the indieauth-metadata URL. e.g. for
	// an indieauth-metadata endpoint
	// https://example.com/.well-known/oauth-authorization-server, the
	// issuer URL could be https://example.com/, or for a metadata URL of
	// https://example.com/wp-json/indieauth/1.0/metadata, the issuer URL
	// could be https://example.com/wp-json/indieauth/1.0
	Issuer *ClientID

	// The Authorization Endpoint.
	AuthorizationEndpoint *url.URL

	// The Token Endpoint.
	TokenEndpoint *url.URL

	// The Ticket Endpoint.
	TicketEndpoint *url.URL

	// The Micropub Endpoint.
	MicropubEndpoint *url.URL

	// The Microsub Endpoint.
	MicrosubEndpoint *url.URL

	// The Introspection Endpoint.
	IntrospectionEndpoint *url.URL

	// The Revocation Endpoint.
	RevocationEndpoint *url.URL

	// The User Info Endpoint.
	UserinfoEndpoint *url.URL

	// URL of a page containing human-readable information that developers
	// might need to know when using the server. This might be a link to the
	// IndieAuth spec or something more personal to your implementation.
	ServiceDocumentation *url.URL

	// JSON array containing scope values supported by the IndieAuth server.
	// Servers MAY choose not to advertise some supported scope values even
	// when this parameter is used.
	ScopesSupported Scopes

	// JSON array containing the response_type values supported. This
	// differs from RFC8414 in that this parameter is OPTIONAL and that, if
	// omitted, the default is code.
	ResponseTypesSupported []ResponseType

	// JSON array containing grant type values supported. If omitted, the
	// default value differs from RFC8414 and is authorization_code.
	GrantTypesSupported []GrantType

	// JSON array containing the methods supported for PKCE. This parameter
	// parameter differs from RFC8414 in that it is not optional as PKCE is
	// REQUIRED.
	CodeChallengeMethodsSupported []CodeChallengeMethod

	// List of client authentication methods supported by this introspection endpoint.
	IntrospectionEndpointAuthMethodsSupported []string // ["Bearer"]

	RevocationEndpointAuthMethodsSupported []string // ["none"]

	// Boolean parameter indicating whether the authorization server
	// provides the iss parameter. If omitted, the default value is false.
	// As the iss parameter is REQUIRED, this is provided for compatibility
	// with OAuth 2.0 servers implementing the parameter.
	AuthorizationResponseIssParameterSupported bool
}

// TestMetadata returns valid random generated Metadata for tests.
func TestMetadata(tb testing.TB) *Metadata {
	tb.Helper()

	return &Metadata{
		Issuer:                TestClientID(tb),
		AuthorizationEndpoint: &url.URL{Scheme: "https", Host: "indieauth.example.com", Path: "/auth"},
		TokenEndpoint:         &url.URL{Scheme: "https", Host: "indieauth.example.com", Path: "/token"},
		TicketEndpoint:        &url.URL{Scheme: "https", Host: "auth.example.org", Path: "/ticket"},
		MicropubEndpoint:      &url.URL{Scheme: "https", Host: "micropub.example.com", Path: "/"},
		MicrosubEndpoint:      &url.URL{Scheme: "https", Host: "microsub.example.com", Path: "/"},
		IntrospectionEndpoint: &url.URL{Scheme: "https", Host: "indieauth.example.com", Path: "/introspect"},
		RevocationEndpoint:    &url.URL{Scheme: "https", Host: "indieauth.example.com", Path: "/revocation"},
		UserinfoEndpoint:      &url.URL{Scheme: "https", Host: "indieauth.example.com", Path: "/userinfo"},
		ServiceDocumentation:  &url.URL{Scheme: "https", Host: "indieauth.net", Path: "/draft/"},
		ScopesSupported: Scopes{
			ScopeBlock,
			ScopeChannels,
			ScopeCreate,
			ScopeDelete,
			ScopeDraft,
			ScopeEmail,
			ScopeFollow,
			ScopeMedia,
			ScopeMute,
			ScopeProfile,
			ScopeRead,
			ScopeUpdate,
		},
		ResponseTypesSupported: []ResponseType{
			ResponseTypeCode,
			ResponseTypeID,
		},
		GrantTypesSupported: []GrantType{
			GrantTypeAuthorizationCode,
			GrantTypeTicket,
		},
		CodeChallengeMethodsSupported: []CodeChallengeMethod{
			CodeChallengeMethodMD5,
			CodeChallengeMethodPLAIN,
			CodeChallengeMethodS1,
			CodeChallengeMethodS256,
			CodeChallengeMethodS512,
		},
		IntrospectionEndpointAuthMethodsSupported:  []string{"Bearer"},
		RevocationEndpointAuthMethodsSupported:     []string{"none"},
		AuthorizationResponseIssParameterSupported: true,
	}
}
