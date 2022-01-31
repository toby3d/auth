package domain

//nolint: tagliatelle
type Metadata struct {
	// The server's issuer identifier. The issuer identifier is a URL that
	// uses the "https" scheme and has no query or fragment components. The
	// identifier MUST be a prefix of the indieauth-metadata URL. e.g. for
	// an indieauth-metadata endpoint
	// https://example.com/.well-known/oauth-authorization-server, the
	// issuer URL could be https://example.com/, or for a metadata URL of
	// https://example.com/wp-json/indieauth/1.0/metadata, the issuer URL
	// could be https://example.com/wp-json/indieauth/1.0
	Issuer *URL `json:"issuer"`

	// The Authorization Endpoint.
	AuthorizationEndpoint *URL `json:"authorization_endpoint"`

	// The Token Endpoint.
	TokenEndpoint *URL `json:"token_endpoint"`

	// JSON array containing scope values supported by the IndieAuth server.
	// Servers MAY choose not to advertise some supported scope values even
	// when this parameter is used.
	ScopesSupported Scopes `json:"scopes_supported,omitempty"`

	// JSON array containing the response_type values supported. This
	// differs from RFC8414 in that this parameter is OPTIONAL and that, if
	// omitted, the default is code.
	ResponseTypesSupported []ResponseType `json:"response_types_supported,omitempty"`

	// JSON array containing grant type values supported. If omitted, the
	// default value differs from RFC8414 and is authorization_code.
	GrantTypesSupported []GrantType `json:"grant_types_supported,omitempty"`

	// URL of a page containing human-readable information that developers
	// might need to know when using the server. This might be a link to the
	// IndieAuth spec or something more personal to your implementation.
	ServiceDocumentation *URL `json:"service_documentation,omitempty"`

	// JSON array containing the methods supported for PKCE. This parameter
	// parameter differs from RFC8414 in that it is not optional as PKCE is
	// REQUIRED.
	CodeChallengeMethodsSupported []CodeChallengeMethod `json:"code_challenge_methods_supported"`

	// Boolean parameter indicating whether the authorization server
	// provides the iss parameter. If omitted, the default value is false.
	// As the iss parameter is REQUIRED, this is provided for compatibility
	// with OAuth 2.0 servers implementing the parameter.
	AuthorizationResponseIssParameterSupported bool `json:"authorization_response_iss_parameter_supported,omitempty"` //nolint: lll

	// The Ticket Endpoint.
	// WARN(toby3d): experimental
	TicketEndpoint *URL `json:"ticket_endpoint,omitempty"`

	// The Micropub Endpoint.
	// WARN(toby3d): experimental
	Micropub *URL `json:"micropub,omitempty"`

	// The Microsub Endpoint.
	// WARN(toby3d): experimental
	Microsub *URL `json:"microsub,omitempty"`
}
