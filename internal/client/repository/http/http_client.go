package http

import (
	"context"
	"fmt"

	http "github.com/valyala/fasthttp"

	"source.toby3d.me/website/indieauth/internal/client"
	"source.toby3d.me/website/indieauth/internal/domain"
	"source.toby3d.me/website/indieauth/internal/util"
)

type httpClientRepository struct {
	client *http.Client
}

const DefaultMaxRedirectsCount int = 10

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

	if err := repo.client.DoRedirects(req, resp, DefaultMaxRedirectsCount); err != nil {
		return nil, fmt.Errorf("failed to make a request to the client: %w", err)
	}

	if resp.StatusCode() == http.StatusNotFound {
		return nil, client.ErrNotExist
	}

	client := &domain.Client{
		ID:          id,
		RedirectURI: make([]*domain.URL, 0),
		Logo:        make([]*domain.URL, 0),
		URL:         make([]*domain.URL, 0),
		Name:        make([]string, 0),
	}

	extract(client, resp)

	return client, nil
}

func extract(dst *domain.Client, src *http.Response) {
	for _, u := range util.ExtractEndpoints(src, relRedirectURI) {
		if containsURL(dst.RedirectURI, u) {
			continue
		}

		dst.RedirectURI = append(dst.RedirectURI, u)
	}

	for _, t := range []string{hXApp, hApp} {
		for _, name := range util.ExtractProperty(src, t, propertyName) {
			n, ok := name.(string)
			if !ok || containsString(dst.Name, n) {
				continue
			}

			dst.Name = append(dst.Name, n)
		}

		for _, logo := range util.ExtractProperty(src, t, propertyLogo) {
			var err error

			var u *domain.URL
			switch l := logo.(type) {
			case string:
				u, err = domain.ParseURL(l)
			case map[string]string:
				value, ok := l["value"]
				if !ok {
					continue
				}

				u, err = domain.ParseURL(value)
			}

			if err != nil {
				continue
			}

			if containsURL(dst.Logo, u) {
				continue
			}

			dst.Logo = append(dst.Logo, u)
		}

		for _, url := range util.ExtractProperty(src, t, propertyURL) {
			l, ok := url.(string)
			if !ok {
				continue
			}

			u, err := domain.ParseURL(l)
			if err != nil {
				continue
			}

			if containsURL(dst.URL, u) {
				continue
			}

			dst.URL = append(dst.URL, u)
		}
	}
}

func containsString(src []string, find string) bool {
	for i := range src {
		if src[i] != find {
			continue
		}

		return true
	}

	return false
}

func containsURL(src []*domain.URL, find *domain.URL) bool {
	for i := range src {
		if src[i].String() != find.String() {
			continue
		}

		return true
	}

	return false
}
