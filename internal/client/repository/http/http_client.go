package http

import (
	"context"
	"fmt"
	"net"
	"net/url"

	http "github.com/valyala/fasthttp"

	"source.toby3d.me/toby3d/auth/internal/client"
	"source.toby3d.me/toby3d/auth/internal/domain"
	"source.toby3d.me/toby3d/auth/internal/httputil"
)

type httpClientRepository struct {
	client *http.Client
}

const (
	DefaultMaxRedirectsCount int = 10

	hApp           string = "h-app"
	hXApp          string = "h-x-app"
	propertyLogo   string = "logo"
	propertyName   string = "name"
	propertyURL    string = "url"
	relRedirectURI string = "redirect_uri"
)

func NewHTTPClientRepository(c *http.Client) client.Repository {
	return &httpClientRepository{
		client: c,
	}
}

func (repo *httpClientRepository) Get(ctx context.Context, cid *domain.ClientID) (*domain.Client, error) {
	ips, err := net.LookupIP(cid.URL().Hostname())
	if err != nil {
		return nil, fmt.Errorf("cannot resolve client IP by id: %w", err)
	}

	for _, ip := range ips {
		if !ip.IsLoopback() {
			continue
		}

		return nil, client.ErrNotExist
	}

	req := http.AcquireRequest()
	defer http.ReleaseRequest(req)
	req.SetRequestURI(cid.String())
	req.Header.SetMethod(http.MethodGet)

	resp := http.AcquireResponse()
	defer http.ReleaseResponse(resp)

	if err := repo.client.DoRedirects(req, resp, DefaultMaxRedirectsCount); err != nil {
		return nil, fmt.Errorf("failed to make a request to the client: %w", err)
	}

	if resp.StatusCode() == http.StatusNotFound {
		return nil, fmt.Errorf("%w: status on client page is not 200", client.ErrNotExist)
	}

	client := &domain.Client{
		ID:          cid,
		RedirectURI: make([]*url.URL, 0),
		Logo:        make([]*url.URL, 0),
		URL:         make([]*url.URL, 0),
		Name:        make([]string, 0),
	}

	extract(client, resp)

	return client, nil
}

//nolint:gocognit,cyclop
func extract(dst *domain.Client, src *http.Response) {
	for _, endpoint := range httputil.ExtractEndpoints(src, relRedirectURI) {
		if !containsURL(dst.RedirectURI, endpoint) {
			dst.RedirectURI = append(dst.RedirectURI, endpoint)
		}
	}

	for _, itemType := range []string{hXApp, hApp} {
		for _, name := range httputil.ExtractProperty(src, itemType, propertyName) {
			if n, ok := name.(string); ok && !containsString(dst.Name, n) {
				dst.Name = append(dst.Name, n)
			}
		}

		for _, logo := range httputil.ExtractProperty(src, itemType, propertyLogo) {
			var (
				u   *url.URL
				err error
			)

			switch l := logo.(type) {
			case string:
				u, err = url.Parse(l)
			case map[string]string:
				if value, ok := l["value"]; ok {
					u, err = url.Parse(value)
				}
			}

			if err != nil || containsURL(dst.Logo, u) {
				continue
			}

			dst.Logo = append(dst.Logo, u)
		}

		for _, property := range httputil.ExtractProperty(src, itemType, propertyURL) {
			prop, ok := property.(string)
			if !ok {
				continue
			}

			if u, err := url.Parse(prop); err == nil || !containsURL(dst.URL, u) {
				dst.URL = append(dst.URL, u)
			}
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

func containsURL(src []*url.URL, find *url.URL) bool {
	for i := range src {
		if src[i].String() != find.String() {
			continue
		}

		return true
	}

	return false
}
