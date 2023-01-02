package http

import (
	"context"
	"fmt"
	"net/url"

	http "github.com/valyala/fasthttp"

	"source.toby3d.me/toby3d/auth/internal/domain"
	"source.toby3d.me/toby3d/auth/internal/httputil"
	"source.toby3d.me/toby3d/auth/internal/profile"
)

type httpProfileRepository struct {
	client *http.Client
}

const (
	ErrPrefix                string = "http"
	DefaultMaxRedirectsCount int    = 10

	hCard         string = "h-card"
	propertyEmail string = "email"
	propertyName  string = "name"
	propertyPhoto string = "photo"
	propertyURL   string = "url"
)

func NewHTPPClientRepository(client *http.Client) profile.Repository {
	return &httpProfileRepository{
		client: client,
	}
}

//nolint:cyclop
func (repo *httpProfileRepository) Get(ctx context.Context, me *domain.Me) (*domain.Profile, error) {
	req := http.AcquireRequest()
	defer http.ReleaseRequest(req)
	req.Header.SetMethod(http.MethodGet)
	req.SetRequestURI(me.String())

	resp := http.AcquireResponse()
	defer http.ReleaseResponse(resp)

	if err := repo.client.DoRedirects(req, resp, DefaultMaxRedirectsCount); err != nil {
		return nil, fmt.Errorf("%s: cannot fetch user by me: %w", ErrPrefix, err)
	}

	result := domain.NewProfile()

	for _, name := range httputil.ExtractProperty(resp, hCard, propertyName) {
		if n, ok := name.(string); ok {
			result.Name = append(result.Name, n)
		}
	}

	for _, rawEmail := range httputil.ExtractProperty(resp, hCard, propertyEmail) {
		email, ok := rawEmail.(string)
		if !ok {
			continue
		}

		if e, err := domain.ParseEmail(email); err == nil {
			result.Email = append(result.Email, e)
		}
	}

	for _, rawURL := range httputil.ExtractProperty(resp, hCard, propertyURL) {
		rawURL, ok := rawURL.(string)
		if !ok {
			continue
		}

		if u, err := url.Parse(rawURL); err == nil {
			result.URL = append(result.URL, u)
		}
	}

	for _, rawPhoto := range httputil.ExtractProperty(resp, hCard, propertyPhoto) {
		photo, ok := rawPhoto.(string)
		if !ok {
			continue
		}

		if p, err := url.Parse(photo); err == nil {
			result.Photo = append(result.Photo, p)
		}
	}

	if result.GetName() == "" && result.GetURL() == nil &&
		result.GetPhoto() == nil && result.GetEmail() == nil {
		return nil, profile.ErrNotExist
	}

	return result, nil
}
