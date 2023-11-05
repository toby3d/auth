package http

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/goccy/go-json"
	"github.com/tomnomnom/linkheader"
	"golang.org/x/exp/slices"
	"willnorris.com/go/microformats"

	"source.toby3d.me/toby3d/auth/internal/common"
	"source.toby3d.me/toby3d/auth/internal/domain"
	"source.toby3d.me/toby3d/auth/internal/user"
)

type (
	//nolint:tagliatelle,lll
	MetadataResponse struct {
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

	httpUserRepository struct {
		client *http.Client
	}
)

const DefaultMaxRedirectsCount int = 10

func NewHTTPUserRepository(client *http.Client) user.Repository {
	return &httpUserRepository{
		client: client,
	}
}

// WARN(toby3d): not implemented.
func (httpUserRepository) Create(_ context.Context, _ domain.User) error {
	return nil
}

//nolint:funlen
func (repo *httpUserRepository) Get(ctx context.Context, me domain.Me) (*domain.User, error) {
	resp, err := repo.client.Get(me.String())
	if err != nil {
		return nil, fmt.Errorf("cannot fetch user by me: %w", err)
	}

	out := &domain.User{
		Profile: new(domain.Profile),
	}
	// NOTE(toby3d): resolved Me may be different from user-provided Me
	out.Me, _ = domain.ParseMe(resp.Request.URL.String())

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("cannot read response body: %w", err)
	}

	mf2 := microformats.Parse(bytes.NewReader(body), resp.Request.URL)

	// NOTE(toby3d): fetch user profile from nodes
	for i := range mf2.Items {
		if !slices.Contains(mf2.Items[i].Type, common.HCard) {
			continue
		}

		parseProfile(mf2.Items[i].Properties, out.Profile)

		break
	}

	// NOTE(toby3d): fetch endpoints from HTML nodes
	for key, dst := range map[string]**url.URL{
		common.RelAuthorizationEndpoint: &out.AuthorizationEndpoint,
		common.RelIndieAuthMetadata:     &out.IndieAuthMetadata,
		common.RelMicropub:              &out.Micropub,
		common.RelMicrosub:              &out.Microsub,
		common.RelTicketEndpoint:        &out.TicketEndpoint,
		common.RelTokenEndpoint:         &out.TokenEndpoint,
	} {
		vals, ok := mf2.Rels[key]
		if !ok || len(vals) == 0 {
			continue
		}

		for i := range vals {
			var u *url.URL
			if u, err = url.Parse(vals[i]); err == nil {
				*dst = u

				break
			}
		}
	}

	// NOTE(toby3d): fetch endpoints from Link header
	for _, link := range linkheader.Parse(resp.Header.Get(common.HeaderLink)) {
		for key, dst := range map[string]**url.URL{
			common.RelAuthorizationEndpoint: &out.AuthorizationEndpoint,
			common.RelIndieAuthMetadata:     &out.IndieAuthMetadata,
			common.RelMicropub:              &out.Micropub,
			common.RelMicrosub:              &out.Microsub,
			common.RelTicketEndpoint:        &out.TicketEndpoint,
			common.RelTokenEndpoint:         &out.TokenEndpoint,
		} {
			if link.Rel != key {
				continue
			}

			var u *url.URL
			if u, err = url.Parse(link.URL); err == nil {
				*dst = u

				break
			}
		}
	}

	if out.IndieAuthMetadata == nil {
		return out, nil
	}

	// NOTE(toby3d): fetch endpoints from metadata payload
	if resp, err = repo.client.Get(out.IndieAuthMetadata.String()); err != nil {
		return out, fmt.Errorf("cannot fetch endpoints from provided metadata URL: %w", err)
	}

	metadata := new(MetadataResponse)
	if err = json.NewDecoder(resp.Body).Decode(metadata); err != nil {
		return out, fmt.Errorf("cannot decode metadata response: %w", err)
	}

	for src, dst := range map[domain.URL]**url.URL{
		metadata.AuthorizationEndpoint: &out.AuthorizationEndpoint,
		metadata.Micropub:              &out.Micropub,
		metadata.Microsub:              &out.Microsub,
		metadata.TicketEndpoint:        &out.TicketEndpoint,
		metadata.TokenEndpoint:         &out.TokenEndpoint,
	} {
		if src.URL == nil {
			continue
		}

		*dst = src.URL
	}

	return out, nil
}

func parseProfile(src map[string][]any, dst *domain.Profile) {
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

	for _, val := range src[common.PropertyPhoto] {
		v, ok := val.(string)
		if !ok {
			continue
		}

		var err error
		if dst.Photo, err = url.Parse(v); err != nil {
			continue
		}

		break
	}

	for _, val := range src[common.PropertyEmail] {
		v, ok := val.(string)
		if !ok {
			continue
		}

		var err error
		if dst.Email, err = domain.ParseEmail(v); err != nil {
			continue
		}

		break
	}
}
