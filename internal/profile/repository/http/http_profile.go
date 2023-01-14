package http

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"

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

// WARN(toby3d): not implemented.
func (repo *httpProfileRepository) Create(_ context.Context, _ domain.Me, _ domain.Profile) error {
	return nil
}

//nolint:cyclop
func (repo *httpProfileRepository) Get(ctx context.Context, me domain.Me) (*domain.Profile, error) {
	resp, err := repo.client.Get(me.String())
	if err != nil {
		return nil, fmt.Errorf("%s: cannot fetch user by me: %w", ErrPrefix, err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("cannot read response body: %w", err)
	}

	buf := bytes.NewReader(body)
	result := domain.NewProfile()

	for _, name := range httputil.ExtractProperty(buf, me.URL(), hCard, propertyName) {
		if n, ok := name.(string); ok {
			result.Name = append(result.Name, n)
		}
	}

	for _, rawEmail := range httputil.ExtractProperty(buf, me.URL(), hCard, propertyEmail) {
		email, ok := rawEmail.(string)
		if !ok {
			continue
		}

		if e, err := domain.ParseEmail(email); err == nil {
			result.Email = append(result.Email, e)
		}
	}

	for _, rawURL := range httputil.ExtractProperty(buf, me.URL(), hCard, propertyURL) {
		rawURL, ok := rawURL.(string)
		if !ok {
			continue
		}

		if u, err := url.Parse(rawURL); err == nil {
			result.URL = append(result.URL, u)
		}
	}

	for _, rawPhoto := range httputil.ExtractProperty(buf, me.URL(), hCard, propertyPhoto) {
		photo, ok := rawPhoto.(string)
		if !ok {
			continue
		}

		if p, err := url.Parse(photo); err == nil {
			result.Photo = append(result.Photo, p)
		}
	}

	// TODO(toby3d): create method like result.Empty()?
	if result.GetName() == "" && result.GetURL() == nil && result.GetPhoto() == nil && result.GetEmail() == nil {
		return nil, profile.ErrNotExist
	}

	return result, nil
}
