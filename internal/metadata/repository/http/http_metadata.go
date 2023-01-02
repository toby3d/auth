package http

import (
	"context"
	"encoding/json"
	"fmt"

	http "github.com/valyala/fasthttp"

	"source.toby3d.me/toby3d/auth/internal/domain"
	"source.toby3d.me/toby3d/auth/internal/httputil"
	"source.toby3d.me/toby3d/auth/internal/metadata"
)

type (
	//nolint:tagliatelle,lll
	Metadata struct {
		Issuer                                     *domain.ClientID             `json:"issuer"`
		AuthorizationEndpoint                      *domain.URL                  `json:"authorization_endpoint"`
		IntrospectionEndpoint                      *domain.URL                  `json:"introspection_endpoint"`
		RevocationEndpoint                         *domain.URL                  `json:"revocation_endpoint,omitempty"`
		ServiceDocumentation                       *domain.URL                  `json:"service_documentation,omitempty"`
		TokenEndpoint                              *domain.URL                  `json:"token_endpoint"`
		UserinfoEndpoint                           *domain.URL                  `json:"userinfo_endpoint,omitempty"`
		CodeChallengeMethodsSupported              []domain.CodeChallengeMethod `json:"code_challenge_methods_supported"`
		GrantTypesSupported                        []domain.GrantType           `json:"grant_types_supported,omitempty"`
		ResponseTypesSupported                     []domain.ResponseType        `json:"response_types_supported,omitempty"`
		ScopesSupported                            []domain.Scope               `json:"scopes_supported,omitempty"`
		IntrospectionEndpointAuthMethodsSupported  []string                     `json:"introspection_endpoint_auth_methods_supported,omitempty"`
		RevocationEndpointAuthMethodsSupported     []string                     `json:"revocation_endpoint_auth_methods_supported,omitempty"`
		AuthorizationResponseIssParameterSupported bool                         `json:"authorization_response_iss_parameter_supported,omitempty"`
	}

	httpMetadataRepository struct {
		client *http.Client
	}
)

const DefaultMaxRedirectsCount int = 10

func NewHTTPMetadataRepository(client *http.Client) metadata.Repository {
	return &httpMetadataRepository{
		client: client,
	}
}

func (repo *httpMetadataRepository) Get(ctx context.Context, me *domain.Me) (*domain.Metadata, error) {
	req := http.AcquireRequest()
	defer http.ReleaseRequest(req)
	req.SetRequestURI(me.String())
	req.Header.SetMethod(http.MethodGet)

	resp := http.AcquireResponse()
	defer http.ReleaseResponse(resp)

	if err := repo.client.DoRedirects(req, resp, DefaultMaxRedirectsCount); err != nil {
		return nil, fmt.Errorf("failed to make a request to the client: %w", err)
	}

	endpoints := httputil.ExtractEndpoints(resp, "indieauth-metadata")
	if len(endpoints) == 0 {
		return nil, metadata.ErrNotExist
	}

	_, body, err := repo.client.Get(nil, endpoints[len(endpoints)-1].String())
	if err != nil {
		return nil, fmt.Errorf("failed to fetch metadata endpoint configuration: %w", err)
	}

	data := new(Metadata)
	if err = json.Unmarshal(body, data); err != nil {
		return nil, fmt.Errorf("cannot unmarshal metadata configuration: %w", err)
	}

	//nolint:exhaustivestruct // TODO(toby3d)
	return &domain.Metadata{
		AuthorizationEndpoint:                      data.AuthorizationEndpoint.URL,
		AuthorizationResponseIssParameterSupported: data.AuthorizationResponseIssParameterSupported,
		CodeChallengeMethodsSupported:              data.CodeChallengeMethodsSupported,
		GrantTypesSupported:                        data.GrantTypesSupported,
		Issuer:                                     data.Issuer,
		ResponseTypesSupported:                     data.ResponseTypesSupported,
		ScopesSupported:                            data.ScopesSupported,
		ServiceDocumentation:                       data.ServiceDocumentation.URL,
		TokenEndpoint:                              data.TokenEndpoint.URL,
		// TODO(toby3d): support extensions?
		// Micropub: data.Micropub,
		// Microsub: data.Microsub,
		// TicketEndpoint: data.TicketEndpoint,
	}, nil
}
