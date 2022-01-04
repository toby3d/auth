package http

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/tomnomnom/linkheader"
	http "github.com/valyala/fasthttp"
	"willnorris.com/go/microformats"

	"source.toby3d.me/website/indieauth/internal/client"
	"source.toby3d.me/website/indieauth/internal/domain"
)

type httpClientRepository struct {
	client *http.Client
}

const (
	relRedirectURI string = "redirect_uri"

	hApp  string = "h-app"
	hXApp string = "h-x-app"

	propertyLogo string = "logo"
	propertyName string = "name"
	propertyURL  string = "url"
)

func NewHTTPClientRepository(c *http.Client) client.Repository {
	return &httpClientRepository{
		client: c,
	}
}

func (repo *httpClientRepository) Get(ctx context.Context, id *domain.ClientID) (*domain.Client, error) {
	req := http.AcquireRequest()
	defer http.ReleaseRequest(req)
	req.SetRequestURI(id.String())
	req.Header.SetMethod(http.MethodGet)

	resp := http.AcquireResponse()
	defer http.ReleaseResponse(resp)

	if err := repo.client.Do(req, resp); err != nil {
		return nil, fmt.Errorf("failed to make a request to the client: %w", err)
	}

	if resp.StatusCode() == http.StatusNotFound {
		return nil, client.ErrNotExist
	}

	client := &domain.Client{
		ID:          id,
		Logo:        make([]*domain.URL, 0),
		Name:        extractValues(resp, propertyName),
		RedirectURI: extractEndpoints(resp, relRedirectURI),
		URL:         make([]*domain.URL, 0),
	}

	for _, v := range extractValues(resp, propertyLogo) {
		u, err := domain.NewURL(v)
		if err != nil {
			continue
		}

		client.Logo = append(client.Logo, u)
	}

	for _, v := range extractValues(resp, propertyURL) {
		u, err := domain.NewURL(v)
		if err != nil {
			continue
		}

		client.URL = append(client.URL, u)
	}

	return client, nil
}

func extractEndpoints(resp *http.Response, name string) []*domain.URL {
	results := make([]*domain.URL, 0)
	endpoints, _ := extractEndpointsFromHeader(resp, name)
	results = append(results, endpoints...)
	endpoints, _ = extractEndpointsFromBody(resp, name)
	results = append(results, endpoints...)

	return results
}

func extractValues(resp *http.Response, key string) []string {
	results := make([]string, 0)

	for _, item := range microformats.Parse(bytes.NewReader(resp.Body()), nil).Items {
		if len(item.Type) == 0 || (item.Type[0] != hApp && item.Type[0] != hXApp) {
			continue
		}

		properties, ok := item.Properties[key]
		if !ok || len(properties) == 0 {
			return nil
		}

		for j := range properties {
			switch p := properties[j].(type) {
			case string:
				results = append(results, p)
			case map[string][]interface{}:
				for _, val := range p["value"] {
					v, ok := val.(string)
					if !ok {
						continue
					}

					results = append(results, v)
				}
			}
		}

		return results
	}

	return nil
}

func extractEndpointsFromHeader(resp *http.Response, name string) ([]*domain.URL, error) {
	results := make([]*domain.URL, 0)

	for _, link := range linkheader.Parse(string(resp.Header.Peek(http.HeaderLink))) {
		if !strings.EqualFold(link.Rel, name) {
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

func extractEndpointsFromBody(resp *http.Response, name string) ([]*domain.URL, error) {
	host, err := url.Parse(string(resp.Header.Peek(http.HeaderHost)))
	if err != nil {
		return nil, fmt.Errorf("cannot parse host header: %w", err)
	}

	endpoints, ok := microformats.Parse(bytes.NewReader(resp.Body()), host).Rels[name]
	if !ok || len(endpoints) == 0 {
		return nil, nil
	}

	results := make([]*domain.URL, 0)
	for i := range endpoints {
		u := http.AcquireURI()
		u.Update(endpoints[i])

		results = append(results, &domain.URL{URI: u})
	}

	return results, nil
}
