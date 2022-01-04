package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/tomnomnom/linkheader"
	http "github.com/valyala/fasthttp"
	"willnorris.com/go/microformats"

	"source.toby3d.me/website/indieauth/internal/domain"
	"source.toby3d.me/website/indieauth/internal/ticket"
)

type (
	//nolint: tagliatelle
	Response struct {
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
	}

	httpTicketRepository struct {
		client *http.Client
	}
)

const (
	relIndieAuthMetadata string = "indieauth-metadata"
	relTokenEndpoint     string = "token_endpoint"
)

func NewHTTPTicketRepository(client *http.Client) ticket.Repository {
	return &httpTicketRepository{
		client: client,
	}
}

func (repo *httpTicketRepository) Get(_ context.Context, resource *domain.URL) (*domain.URL, error) {
	req := http.AcquireRequest()
	defer http.ReleaseRequest(req)
	req.Header.SetMethod(http.MethodGet)
	req.SetRequestURIBytes(resource.FullURI())

	resp := http.AcquireResponse()
	defer http.ReleaseResponse(resp)

	if err := repo.client.Do(req, resp); err != nil {
		return nil, fmt.Errorf("cannot fetch resource URL: %w", err)
	}

	if metadataEndpoint := extractEndpoint(resp, relIndieAuthMetadata); metadataEndpoint != nil {
		if result, err := extractFromMetadata(repo.client, metadataEndpoint); err == nil && result != nil {
			return result, nil
		}
	}

	if result := extractEndpoint(resp, relTokenEndpoint); result != nil {
		return result, nil
	}

	return nil, ticket.ErrNotExist
}

func extractEndpoint(resp *http.Response, name string) *domain.URL {
	u, err := extractEndpointFromHeader(resp, name)
	if err == nil && u != nil {
		return u
	}

	if u, err = extractEndpointFromBody(resp, name); err == nil && u != nil {
		return u
	}

	return nil
}

func extractEndpointFromHeader(resp *http.Response, name string) (*domain.URL, error) {
	for _, link := range linkheader.Parse(string(resp.Header.Peek(http.HeaderLink))) {
		if !strings.EqualFold(link.Rel, name) {
			continue
		}

		u := http.AcquireURI()
		if err := u.Parse(resp.Header.Peek(http.HeaderHost), []byte(link.URL)); err != nil {
			return nil, err
		}

		return &domain.URL{URI: u}, nil
	}

	return nil, nil
}

func extractEndpointFromBody(resp *http.Response, name string) (*domain.URL, error) {
	host, err := url.Parse(string(resp.Header.Peek(http.HeaderHost)))
	if err != nil {
		return nil, fmt.Errorf("cannot parse host header: %w", err)
	}

	endpoints, ok := microformats.Parse(bytes.NewReader(resp.Body()), host).Rels[name]
	if !ok || len(endpoints) == 0 {
		return nil, nil
	}

	return domain.NewURL(endpoints[len(endpoints)-1])
}

func extractFromMetadata(client *http.Client, endpoint *domain.URL) (*domain.URL, error) {
	_, body, err := client.Get(nil, endpoint.String())
	if err != nil {
		return nil, err
	}

	resp := new(Response)
	if err = json.Unmarshal(body, resp); err != nil {
		return nil, err
	}

	return resp.TokenEndpoint, nil
}
