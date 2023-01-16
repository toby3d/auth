package http

//nolint:tagliatelle // https://indieauth.net/source/#indieauth-server-metadata
type MetadataResponse struct {
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
	IntrospectionEndpointAuthMethodsSupported []string `json:"introspection_endpoint_auth_methods_supported,omitempty"` //nolint:lll

	// The Revocation Endpoint.
	RevocationEndpoint string `json:"revocation_endpoint,omitempty"`

	// JSON array containing the value "none".
	RevocationEndpointAuthMethodsSupported []string `json:"revocation_endpoint_auth_methods_supported,omitempty"` //nolint:lll

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
	AuthorizationResponseIssParameterSupported bool `json:"authorization_response_iss_parameter_supported,omitempty"` //nolint:lll

	// The User Info Endpoint.
	UserinfoEndpoint string `json:"userinfo_endpoint,omitempty"`
}
