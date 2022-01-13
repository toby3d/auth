package util

import (
	"bytes"
	"encoding/json"
	"net/url"
	"strings"

	"github.com/tomnomnom/linkheader"
	http "github.com/valyala/fasthttp"
	"willnorris.com/go/microformats"

	"source.toby3d.me/website/indieauth/internal/domain"
)

//nolint: tagliatelle
type Metadata struct {
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
	Issuer *domain.URL `json:"issuer"`

	// The Authorization Endpoint.
	AuthorizationEndpoint *domain.URL `json:"authorization_endpoint"`

	// The Token Endpoint.
	TokenEndpoint *domain.URL `json:"token_endpoint"`

	//  JSON array containing scope values supported by the
	// IndieAuth server. Servers MAY choose not to advertise some
	// supported scope values even when this parameter is used.
	ScopesSupported domain.Scopes `json:"scopes_supported,omitempty"`

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

	// WARN(toby3d): experimental
	// The Ticket Endpoint.
	TicketEndpoint *domain.URL `json:"ticket_endpoint,omitempty"`

	// The Micropub Endpoint.
	Micropub *domain.URL `json:"micropub,omitempty"`

	// The Microsub Endpoint.
	Microsub *domain.URL `json:"microsub,omitempty"`
}

func ExtractEndpoints(resp *http.Response, rel string) []*domain.URL {
	results := make([]*domain.URL, 0)

	urls, err := ExtractEndpointsFromHeader(resp, rel)
	if err == nil {
		results = append(results, urls...)
	}

	urls, err = ExtractEndpointsFromBody(resp, rel)
	if err == nil {
		results = append(results, urls...)
	}

	return results
}

func ExtractEndpointsFromHeader(resp *http.Response, rel string) ([]*domain.URL, error) {
	results := make([]*domain.URL, 0)

	for _, link := range linkheader.Parse(string(resp.Header.Peek(http.HeaderLink))) {
		if !strings.EqualFold(link.Rel, rel) {
			continue
		}

		u := http.AcquireURI()
		if err := u.Parse(resp.Header.Peek(http.HeaderHost), []byte(link.URL)); err != nil {
			return nil, err
		}

		results = append(results, &domain.URL{URI: u})
	}

	return results, nil
}

func ExtractEndpointsFromBody(resp *http.Response, rel string) ([]*domain.URL, error) {
	endpoints, ok := microformats.Parse(bytes.NewReader(resp.Body()), nil).Rels[rel]
	if !ok || len(endpoints) == 0 {
		return nil, nil
	}

	results := make([]*domain.URL, 0)

	for i := range endpoints {
		u := http.AcquireURI()
		if err := u.Parse(resp.Header.Peek(http.HeaderHost), []byte(endpoints[i])); err != nil {
			return nil, err
		}

		results = append(results, &domain.URL{URI: u})
	}

	return results, nil
}

func ExtractMetadata(resp *http.Response, client *http.Client) (*Metadata, error) {
	endpoints := ExtractEndpoints(resp, "indieauth-metadata")
	if endpoints == nil || len(endpoints) == 0 {
		return nil, nil
	}

	_, body, err := client.Get(nil, endpoints[len(endpoints)-1].String())
	if err != nil {
		return nil, err
	}

	result := new(Metadata)
	if err = json.Unmarshal(body, result); err != nil {
		return nil, err
	}

	return result, nil
}

func ExtractProperty(resp *http.Response, t, key string) []interface{} {
	data := microformats.Parse(bytes.NewReader(resp.Body()), &url.URL{
		Host: string(resp.Header.Peek(http.HeaderHost)),
	})

	for _, item := range data.Items {
		if !contains(item.Type, t) {
			continue
		}

		return item.Properties[key]
	}

	return nil
}

func contains(src []string, find string) bool {
	for i := range src {
		if !strings.EqualFold(src[i], find) {
			continue
		}

		return true
	}

	return false
}
