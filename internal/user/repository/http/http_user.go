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
	resolvedMe, _ := domain.ParseMe(string(resp.Header.Peek(http.HeaderLocation)))

	user := &domain.User{
		AuthorizationEndpoint: nil,
		IndieAuthMetadata:     nil,
		Me:                    resolvedMe,
		Micropub:              nil,
		Microsub:              nil,
		Profile:               domain.NewProfile(),
		TicketEndpoint:        nil,
		TokenEndpoint:         nil,
	}

	if metadata, err := util.ExtractMetadata(resp, repo.client); err == nil {
		user.AuthorizationEndpoint = metadata.AuthorizationEndpoint
		user.Micropub = metadata.MicropubEndpoint
		user.Microsub = metadata.MicrosubEndpoint
		user.TicketEndpoint = metadata.TicketEndpoint
		user.TokenEndpoint = metadata.TokenEndpoint
	}

	extractUser(user, resp)
	extractProfile(user.Profile, resp)

	return user, nil
}

//nolint: cyclop
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

//nolint: cyclop
func extractProfile(dst *domain.Profile, src *http.Response) {
	for _, name := range util.ExtractProperty(src, hCard, propertyName) {
		if n, ok := name.(string); ok {
			dst.Name = append(dst.Name, n)
		}
	}

	for _, rawEmail := range util.ExtractProperty(src, hCard, propertyEmail) {
		email, ok := rawEmail.(string)
		if !ok {
			continue
		}

		if e, err := domain.ParseEmail(email); err == nil {
			dst.Email = append(dst.Email, e)
		}
	}

	for _, rawURL := range util.ExtractProperty(src, hCard, propertyURL) {
		url, ok := rawURL.(string)
		if !ok {
			continue
		}

		if u, err := domain.ParseURL(url); err == nil {
			dst.URL = append(dst.URL, u)
		}
	}

	for _, rawPhoto := range util.ExtractProperty(src, hCard, propertyPhoto) {
		photo, ok := rawPhoto.(string)
		if !ok {
			continue
		}

		if p, err := domain.ParseURL(photo); err == nil {
			dst.Photo = append(dst.Photo, p)
		}
	}
}
