package http

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"golang.org/x/exp/slices"

	"source.toby3d.me/toby3d/auth/internal/client"
	"source.toby3d.me/toby3d/auth/internal/common"
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

// WARN(toby3d): not implemented.
func (httpClientRepository) Create(_ context.Context, _ domain.Client) error {
	return nil
}

func (repo httpClientRepository) Get(ctx context.Context, cid domain.ClientID) (*domain.Client, error) {
	resp, err := repo.client.Get(cid.String())
	if err != nil {
		return nil, fmt.Errorf("failed to make a request to the client: %w", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("%w: status on client page is not 200", client.ErrNotExist)
	}

	client := &domain.Client{
		ID:          cid,
		RedirectURI: make([]*url.URL, 0),
		Logo:        make([]*url.URL, 0),
		URL:         make([]*url.URL, 0),
		Name:        make([]string, 0),
	}

	extract(resp.Body, resp.Request.URL, client, resp.Header.Get(common.HeaderLink))

	return client, nil
}

//nolint:gocognit,cyclop
func extract(r io.Reader, u *url.URL, dst *domain.Client, header string) {
	body, _ := io.ReadAll(r)

	for _, endpoint := range httputil.ExtractEndpoints(bytes.NewReader(body), u, header, relRedirectURI) {
		if !containsUrl(dst.RedirectURI, endpoint) {
			dst.RedirectURI = append(dst.RedirectURI, endpoint)
		}
	}

	for _, itemType := range []string{hApp, hXApp} {
		for _, name := range httputil.ExtractProperty(bytes.NewReader(body), u, itemType, propertyName) {
			if n, ok := name.(string); ok && !slices.Contains(dst.Name, n) {
				dst.Name = append(dst.Name, n)
			}
		}

		for _, logo := range httputil.ExtractProperty(bytes.NewReader(body), u, itemType, propertyLogo) {
			var (
				logoURL *url.URL
				err     error
			)

			switch l := logo.(type) {
			case string:
				logoURL, err = url.Parse(l)
			case map[string]string:
				if value, ok := l["value"]; ok {
					logoURL, err = url.Parse(value)
				}
			}

			if err != nil || containsUrl(dst.Logo, logoURL) {
				continue
			}

			dst.Logo = append(dst.Logo, logoURL)
		}

		for _, property := range httputil.ExtractProperty(bytes.NewReader(body), u, itemType, propertyURL) {
			prop, ok := property.(string)
			if !ok {
				continue
			}

			if u, err := url.Parse(prop); err == nil && !containsUrl(dst.URL, u) {
				dst.URL = append(dst.URL, u)
			}
		}
	}
}

func containsUrl(src []*url.URL, find *url.URL) bool {
	for i := range src {
		if src[i].String() != find.String() {
			continue
		}

		return true
	}

	return false
}
