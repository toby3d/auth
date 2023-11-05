package http

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/tomnomnom/linkheader"
	"golang.org/x/exp/slices"
	"willnorris.com/go/microformats"

	"source.toby3d.me/toby3d/auth/internal/client"
	"source.toby3d.me/toby3d/auth/internal/common"
	"source.toby3d.me/toby3d/auth/internal/domain"
)

type (
	//nolint:tagliatelle,lll
	Response struct {
		TicketEndpoint                             domain.URL                   `json:"ticket_endpoint"`
		AuthorizationEndpoint                      domain.URL                   `json:"authorization_endpoint"`
		IntrospectionEndpoint                      domain.URL                   `json:"introspection_endpoint"`
		RevocationEndpoint                         domain.URL                   `json:"revocation_endpoint,omitempty"`
		ServiceDocumentation                       domain.URL                   `json:"service_documentation,omitempty"`
		TokenEndpoint                              domain.URL                   `json:"token_endpoint"`
		UserinfoEndpoint                           domain.URL                   `json:"userinfo_endpoint,omitempty"`
		Microsub                                   domain.URL                   `json:"microsub"`
		Issuer                                     domain.URL                   `json:"issuer"`
		Micropub                                   domain.URL                   `json:"micropub"`
		GrantTypesSupported                        []domain.GrantType           `json:"grant_types_supported,omitempty"`
		IntrospectionEndpointAuthMethodsSupported  []string                     `json:"introspection_endpoint_auth_methods_supported,omitempty"`
		RevocationEndpointAuthMethodsSupported     []string                     `json:"revocation_endpoint_auth_methods_supported,omitempty"`
		ScopesSupported                            []domain.Scope               `json:"scopes_supported,omitempty"`
		ResponseTypesSupported                     []domain.ResponseType        `json:"response_types_supported,omitempty"`
		CodeChallengeMethodsSupported              []domain.CodeChallengeMethod `json:"code_challenge_methods_supported"`
		AuthorizationResponseIssParameterSupported bool                         `json:"authorization_response_iss_parameter_supported,omitempty"`
	}

	httpClientRepository struct {
		client *http.Client
	}
)

func NewHTTPClientRepository(c *http.Client) client.Repository {
	return &httpClientRepository{
		client: c,
	}
}

// WARN(toby3d): not implemented.
func (httpClientRepository) Create(_ context.Context, _ domain.Client) error {
	return nil
}

func (repo httpClientRepository) Get(ctx context.Context, cid domain.ClientID) (*domain.Client, error) {
	out := &domain.Client{
		ID:          cid,
		RedirectURI: make([]*url.URL, 0),
		Logo:        nil,
		URL:         nil,
		Name:        "",
	}

	if cid.IsLocalhost() {
		return out, nil
	}

	resp, err := repo.client.Get(cid.String())
	if err != nil {
		return nil, fmt.Errorf("failed to make a request to the client: %w", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("%w: status on client page is not 200", client.ErrNotExist)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("cannot read response body: %w", err)
	}

	// NOTE(toby3d): fetch redirect uri's and application profile from HTML nodes
	mf2 := microformats.Parse(bytes.NewReader(body), resp.Request.URL)

	for i := range mf2.Items {
		if !slices.Contains(mf2.Items[i].Type, common.HApp) &&
			!slices.Contains(mf2.Items[i].Type, common.HXApp) {
			continue
		}

		parseProfile(mf2.Items[i].Properties, out)
	}

	for _, val := range mf2.Rels[common.RelRedirectURI] {
		var u *url.URL
		if u, err = url.Parse(val); err == nil {
			out.RedirectURI = append(out.RedirectURI, u)
		}
	}

	// NOTE(toby3d): fetch redirect uri's from Link header
	for _, link := range linkheader.Parse(resp.Header.Get(common.HeaderLink)) {
		if link.Rel != common.RelRedirectURI {
			continue
		}

		var u *url.URL
		if u, err = url.Parse(link.URL); err == nil {
			out.RedirectURI = append(out.RedirectURI, u)
		}
	}

	return out, nil
}

func parseProfile(src map[string][]any, dst *domain.Client) {
	for _, val := range src[common.PropertyName] {
		v, ok := val.(string)
		if !ok {
			continue
		}

		dst.Name = v

		break
	}

	for _, val := range src[common.PropertyURL] {
		v, ok := val.(string)
		if !ok {
			continue
		}

		var err error
		if dst.URL, err = url.Parse(v); err != nil {
			continue
		}

		break
	}

	for _, val := range src[common.PropertyLogo] {
		v, ok := val.(string)
		if !ok {
			continue
		}

		var err error
		if dst.Logo, err = url.Parse(v); err != nil {
			continue
		}

		break
	}
}
