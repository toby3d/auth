package http

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/goccy/go-json"
	"github.com/tomnomnom/linkheader"
	"willnorris.com/go/microformats"

	"source.toby3d.me/toby3d/auth/internal/common"
	"source.toby3d.me/toby3d/auth/internal/domain"
	"source.toby3d.me/toby3d/auth/internal/metadata"
)

type (
	//nolint:tagliatelle,lll
	Response struct {
		Issuer                                     domain.URL                   `json:"issuer"`
		AuthorizationEndpoint                      domain.URL                   `json:"authorization_endpoint"`
		IntrospectionEndpoint                      domain.URL                   `json:"introspection_endpoint"`
		RevocationEndpoint                         domain.URL                   `json:"revocation_endpoint,omitempty"`
		ServiceDocumentation                       domain.URL                   `json:"service_documentation,omitempty"`
		TokenEndpoint                              domain.URL                   `json:"token_endpoint"`
		UserinfoEndpoint                           domain.URL                   `json:"userinfo_endpoint,omitempty"`
		CodeChallengeMethodsSupported              []domain.CodeChallengeMethod `json:"code_challenge_methods_supported"`
		GrantTypesSupported                        []domain.GrantType           `json:"grant_types_supported,omitempty"`
		ResponseTypesSupported                     []domain.ResponseType        `json:"response_types_supported,omitempty"`
		ScopesSupported                            []domain.Scope               `json:"scopes_supported,omitempty"`
		IntrospectionEndpointAuthMethodsSupported  []string                     `json:"introspection_endpoint_auth_methods_supported,omitempty"`
		RevocationEndpointAuthMethodsSupported     []string                     `json:"revocation_endpoint_auth_methods_supported,omitempty"`
		AuthorizationResponseIssParameterSupported bool                         `json:"authorization_response_iss_parameter_supported,omitempty"`

		// Extensions
		TicketEndpoint domain.URL `json:"ticket_endpoint"`
		Micropub       domain.URL `json:"micropub"`
		Microsub       domain.URL `json:"microsub"`
	}

	httpMetadataRepository struct {
		client *http.Client
	}
)

const relIndieauthMetadata = "indieauth-metadata"

func NewHTTPMetadataRepository(client *http.Client) metadata.Repository {
	return &httpMetadataRepository{
		client: client,
	}
}

// WARN(toby3d): not implemented.
func (httpMetadataRepository) Create(_ context.Context, _ *url.URL, _ domain.Metadata) error {
	return nil
}

func (repo *httpMetadataRepository) Get(_ context.Context, u *url.URL) (*domain.Metadata, error) {
	resp, err := repo.client.Get(u.String())
	if err != nil {
		return nil, fmt.Errorf("cannot make request to provided Me: %w", err)
	}

	relVals := make(map[string][]string)
	for _, link := range linkheader.Parse(resp.Header.Get(common.HeaderLink)) {
		populateBuffer(relVals, link.Rel, link.URL)
	}

	if mf2 := microformats.Parse(resp.Body, resp.Request.URL); mf2 != nil {
		for rel, vals := range mf2.Rels {
			if len(vals) > 0 {
				populateBuffer(relVals, rel, vals[0])
			}
		}
	}

	out := new(domain.Metadata)
	// NOTE(toby3d): fetch all from metadata endpoint if exists
	if endpoints, ok := relVals["indieauth-metadata"]; ok {
		if resp, err = repo.client.Get(endpoints[0]); err != nil {
			return nil, fmt.Errorf("cannot fetch indieauth-metadata endpoint: %w", err)
		}

		in := NewResponse()
		if err = in.bind(resp); err != nil {
			return nil, err
		}

		in.populate(out)

		return out, nil
	}

	// NOTE(toby3d): metadata not exists, fallback for old clients
	for key, dst := range map[string]**url.URL{
		"authorization_endpoint": &out.AuthorizationEndpoint,
		"micropub":               &out.MicropubEndpoint,
		"microsub":               &out.MicrosubEndpoint,
		"ticket_endpoint":        &out.TicketEndpoint,
		"token_endpoint":         &out.TokenEndpoint,
	} {
		if values, ok := relVals[key]; ok && len(values) > 0 {
			if u, err := url.Parse(values[0]); err == nil {
				*dst = resp.Request.URL.ResolveReference(u)
			}
		}
	}

	return out, nil
}

func populateBuffer(dst map[string][]string, rel, u string) {
	if _, ok := dst[rel]; !ok {
		dst[rel] = make([]string, 0)
	}

	dst[rel] = append(dst[rel], u)
}

func NewResponse() *Response {
	return &Response{
		CodeChallengeMethodsSupported:             make([]domain.CodeChallengeMethod, 0),
		GrantTypesSupported:                       make([]domain.GrantType, 0),
		ResponseTypesSupported:                    make([]domain.ResponseType, 0),
		ScopesSupported:                           make([]domain.Scope, 0),
		IntrospectionEndpointAuthMethodsSupported: make([]string, 0),
		RevocationEndpointAuthMethodsSupported:    make([]string, 0),
	}
}

func (r *Response) bind(resp *http.Response) error {
	if err := json.NewDecoder(resp.Body).Decode(r); err != nil {
		return fmt.Errorf("cannot unmarshal metadata configuration: %w", err)
	}

	return nil
}

func (r Response) populate(dst *domain.Metadata) {
	dst.AuthorizationEndpoint = r.AuthorizationEndpoint.URL
	dst.AuthorizationResponseIssParameterSupported = r.AuthorizationResponseIssParameterSupported
	dst.IntrospectionEndpoint = r.IntrospectionEndpoint.URL
	dst.Issuer = r.Issuer.URL
	dst.MicropubEndpoint = r.Micropub.URL
	dst.MicrosubEndpoint = r.Microsub.URL
	dst.RevocationEndpoint = r.RevocationEndpoint.URL
	dst.ServiceDocumentation = r.ServiceDocumentation.URL
	dst.TicketEndpoint = r.TicketEndpoint.URL
	dst.TokenEndpoint = r.TokenEndpoint.URL
	dst.UserinfoEndpoint = r.UserinfoEndpoint.URL

	for _, scope := range r.ScopesSupported {
		dst.ScopesSupported = append(dst.ScopesSupported, scope)
	}

	for _, method := range r.RevocationEndpointAuthMethodsSupported {
		dst.RevocationEndpointAuthMethodsSupported = append(dst.RevocationEndpointAuthMethodsSupported, method)
	}

	for _, responseType := range r.ResponseTypesSupported {
		dst.ResponseTypesSupported = append(dst.ResponseTypesSupported, responseType)
	}

	for _, method := range r.IntrospectionEndpointAuthMethodsSupported {
		dst.IntrospectionEndpointAuthMethodsSupported = append(dst.IntrospectionEndpointAuthMethodsSupported,
			method)
	}

	for _, grantType := range r.GrantTypesSupported {
		dst.GrantTypesSupported = append(dst.GrantTypesSupported, grantType)
	}

	for _, method := range r.CodeChallengeMethodsSupported {
		dst.CodeChallengeMethodsSupported = append(dst.CodeChallengeMethodsSupported, method)
	}
}
