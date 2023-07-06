package http

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"golang.org/x/exp/slices"

	"source.toby3d.me/toby3d/auth/internal/common"
	"source.toby3d.me/toby3d/auth/internal/domain"
	"source.toby3d.me/toby3d/auth/internal/httputil"
	"source.toby3d.me/toby3d/auth/internal/user"
)

type httpUserRepository struct {
	client *http.Client
}

const (
	DefaultMaxRedirectsCount int = 10

	hCard                    string = "h-card"
	propertyEmail            string = "email"
	propertyName             string = "name"
	propertyPhoto            string = "photo"
	propertyURL              string = "url"
	relAuthorizationEndpoint string = "authorization_endpoint"
	relIndieAuthMetadata     string = "indieauth-metadata"
	relMicropub              string = "micropub"
	relMicrosub              string = "microsub"
	relTicketEndpoint        string = "ticket_endpoint"
	relTokenEndpoint         string = "token_endpoint"
)

func NewHTTPUserRepository(client *http.Client) user.Repository {
	return &httpUserRepository{
		client: client,
	}
}

// WARN(toby3d): not implemented.
func (httpUserRepository) Create(_ context.Context, _ domain.User) error {
	return nil
}

func (repo *httpUserRepository) Get(ctx context.Context, me domain.Me) (*domain.User, error) {
	resp, err := repo.client.Get(me.String())
	if err != nil {
		return nil, fmt.Errorf("cannot fetch user by me: %w", err)
	}

	user := &domain.User{
		AuthorizationEndpoint: nil,
		IndieAuthMetadata:     nil,
		Me:                    &me,
		Micropub:              nil,
		Microsub:              nil,
		Profile:               domain.NewProfile(),
		TicketEndpoint:        nil,
		TokenEndpoint:         nil,
	}

	var metadata *domain.Metadata
	if metadata, err = httputil.ExtractFromMetadata(repo.client, me.String()); err == nil {
		user.AuthorizationEndpoint = metadata.AuthorizationEndpoint
		user.Micropub = metadata.MicropubEndpoint
		user.Microsub = metadata.MicrosubEndpoint
		user.TicketEndpoint = metadata.TicketEndpoint
		user.TokenEndpoint = metadata.TokenEndpoint
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("cannot read response body: %w", err)
	}

	extractUser(me.URL(), user, body, resp.Header.Get(common.HeaderLink))
	extractProfile(me.URL(), user.Profile, body)

	return user, nil
}

//nolint:cyclop
func extractUser(u *url.URL, dst *domain.User, body []byte, header string) {
	for key, target := range map[string]**url.URL{
		relAuthorizationEndpoint: &dst.AuthorizationEndpoint,
		relIndieAuthMetadata:     &dst.IndieAuthMetadata,
		relMicropub:              &dst.Micropub,
		relMicrosub:              &dst.Microsub,
		relTicketEndpoint:        &dst.TicketEndpoint,
		relTokenEndpoint:         &dst.TokenEndpoint,
	} {
		if target == nil {
			continue
		}

		if endpoints := httputil.ExtractEndpoints(bytes.NewReader(body), u, header, key); len(endpoints) > 0 {
			*target = endpoints[len(endpoints)-1]
		}
	}
}

//nolint:cyclop
func extractProfile(u *url.URL, dst *domain.Profile, body []byte) {
	for _, name := range httputil.ExtractProperty(bytes.NewReader(body), u, hCard, propertyName) {
		if n, ok := name.(string); ok && !slices.Contains(dst.Name, n) {
			dst.Name = append(dst.Name, n)
		}
	}

	for _, rawEmail := range httputil.ExtractProperty(bytes.NewReader(body), u, hCard, propertyEmail) {
		email, ok := rawEmail.(string)
		if !ok {
			continue
		}

		if e, err := domain.ParseEmail(email); err == nil && !slices.Contains(dst.Email, e) {
			dst.Email = append(dst.Email, e)
		}
	}

	for _, rawURL := range httputil.ExtractProperty(bytes.NewReader(body), u, hCard, propertyURL) {
		rawURL, ok := rawURL.(string)
		if !ok {
			continue
		}

		if parsedURL, err := url.Parse(rawURL); err == nil && !containsUrl(dst.URL, u) {
			dst.URL = append(dst.URL, parsedURL)
		}
	}

	for _, rawPhoto := range httputil.ExtractProperty(bytes.NewReader(body), u, hCard, propertyPhoto) {
		photo, ok := rawPhoto.(string)
		if !ok {
			continue
		}

		if p, err := url.Parse(photo); err == nil && !containsUrl(dst.Photo, p) {
			dst.Photo = append(dst.Photo, p)
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
