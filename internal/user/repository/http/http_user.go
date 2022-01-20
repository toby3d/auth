package http

import (
	"context"
	"fmt"

	http "github.com/valyala/fasthttp"

	"source.toby3d.me/website/indieauth/internal/domain"
	"source.toby3d.me/website/indieauth/internal/user"
	"source.toby3d.me/website/indieauth/internal/util"
)

type httpUserRepository struct {
	client *http.Client
}

const DefaultMaxRedirectsCount int = 10

const (
	relAuthorizationEndpoint string = "authorization_endpoint"
	relIndieAuthMetadata     string = "indieauth-metadata"
	relMicropub              string = "micropub"
	relMicrosub              string = "microsub"
	relTicketEndpoint        string = "ticket_endpoint"
	relTokenEndpoint         string = "token_endpoint"

	hCard string = "h-card"

	propertyEmail string = "email"
	propertyName  string = "name"
	propertyPhoto string = "photo"
	propertyURL   string = "url"
)

func NewHTTPUserRepository(client *http.Client) user.Repository {
	return &httpUserRepository{
		client: client,
	}
}

func (repo *httpUserRepository) Get(ctx context.Context, me *domain.Me) (*domain.User, error) {
	req := http.AcquireRequest()
	defer http.ReleaseRequest(req)
	req.Header.SetMethod(http.MethodGet)
	req.SetRequestURI(me.String())

	resp := http.AcquireResponse()
	defer http.ReleaseResponse(resp)

	if err := repo.client.DoRedirects(req, resp, DefaultMaxRedirectsCount); err != nil {
		return nil, fmt.Errorf("cannot fetch user by me: %w", err)
	}

	// TODO(toby3d): handle error here?
	resolvedMe, _ := domain.NewMe(string(resp.Header.Peek(http.HeaderLocation)))
	u := &domain.User{
		Me: resolvedMe,
		Profile: &domain.Profile{
			Name:  make([]string, 0),
			URL:   make([]*domain.URL, 0),
			Photo: make([]*domain.URL, 0),
			Email: make([]*domain.Email, 0),
		},
	}

	metadata, err := util.ExtractMetadata(resp, repo.client)
	if err == nil && metadata != nil {
		u.AuthorizationEndpoint = metadata.AuthorizationEndpoint
		u.Micropub = metadata.Micropub
		u.Microsub = metadata.Microsub
		u.TicketEndpoint = metadata.TicketEndpoint
		u.TokenEndpoint = metadata.TokenEndpoint
	}

	extractUser(u, resp)
	extractProfile(u.Profile, resp)

	return u, nil
}

func extractUser(dst *domain.User, src *http.Response) {
	if dst.IndieAuthMetadata != nil {
		if endpoints := util.ExtractEndpoints(src, relIndieAuthMetadata); len(endpoints) > 0 {
			dst.IndieAuthMetadata = endpoints[len(endpoints)-1]
		}
	}

	if dst.AuthorizationEndpoint == nil {
		if endpoints := util.ExtractEndpoints(src, relAuthorizationEndpoint); len(endpoints) > 0 {
			dst.AuthorizationEndpoint = endpoints[len(endpoints)-1]
		}
	}

	if dst.Micropub == nil {
		if endpoints := util.ExtractEndpoints(src, relMicropub); len(endpoints) > 0 {
			dst.Micropub = endpoints[len(endpoints)-1]
		}
	}

	if dst.Microsub == nil {
		if endpoints := util.ExtractEndpoints(src, relMicrosub); len(endpoints) > 0 {
			dst.Microsub = endpoints[len(endpoints)-1]
		}
	}

	if dst.TicketEndpoint == nil {
		if endpoints := util.ExtractEndpoints(src, relTicketEndpoint); len(endpoints) > 0 {
			dst.TicketEndpoint = endpoints[len(endpoints)-1]
		}
	}

	if dst.TokenEndpoint == nil {
		if endpoints := util.ExtractEndpoints(src, relTokenEndpoint); len(endpoints) > 0 {
			dst.TokenEndpoint = endpoints[len(endpoints)-1]
		}
	}
}

func extractProfile(dst *domain.Profile, src *http.Response) {
	for _, name := range util.ExtractProperty(src, hCard, propertyName) {
		n, ok := name.(string)
		if !ok {
			continue
		}

		dst.Name = append(dst.Name, n)
	}

	for _, rawEmail := range util.ExtractProperty(src, hCard, propertyEmail) {
		email, ok := rawEmail.(string)
		if !ok {
			continue
		}

		e, err := domain.NewEmail(email)
		if err != nil {
			continue
		}

		dst.Email = append(dst.Email, e)
	}

	for _, rawUrl := range util.ExtractProperty(src, hCard, propertyURL) {
		url, ok := rawUrl.(string)
		if !ok {
			continue
		}

		u, err := domain.NewURL(url)
		if err != nil {
			continue
		}

		dst.URL = append(dst.URL, u)
	}

	for _, rawPhoto := range util.ExtractProperty(src, hCard, propertyPhoto) {
		photo, ok := rawPhoto.(string)
		if !ok {
			continue
		}

		p, err := domain.NewURL(photo)
		if err != nil {
			continue
		}

		dst.Photo = append(dst.Photo, p)
	}
}
