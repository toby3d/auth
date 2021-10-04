package http

import (
	"bytes"
	"context"
	"net/url"
	"strings"

	"github.com/pkg/errors"
	"github.com/tomnomnom/linkheader"
	http "github.com/valyala/fasthttp"
	"willnorris.com/go/microformats"

	"source.toby3d.me/website/oauth/internal/client"
	"source.toby3d.me/website/oauth/internal/domain"
)

type httpClientRepository struct {
	client *http.Client
}

const (
	HApp  string = "h-app"
	HXApp string = "h-x-app"

	KeyName string = "name"
	KeyLogo string = "logo"
	KeyURL  string = "url"

	ValueValue string = "value"

	RelRedirectURI string = "redirect_uri"
)

func NewHTTPClientRepository(c *http.Client) client.Repository {
	return &httpClientRepository{
		client: c,
	}
}

func (repo *httpClientRepository) Get(ctx context.Context, id string) (*domain.Client, error) {
	req := http.AcquireRequest()
	defer http.ReleaseRequest(req)
	req.SetRequestURI(id)
	req.Header.SetMethod(http.MethodGet)

	resp := http.AcquireResponse()
	defer http.ReleaseResponse(resp)

	if err := repo.client.Do(req, resp); err != nil {
		return nil, errors.Wrap(err, "failed to make a request to the client")
	}

	client := domain.NewClient()
	client.ID = id

	for _, l := range linkheader.Parse(string(resp.Header.Peek(http.HeaderLink))) {
		if !strings.Contains(l.Rel, "redirect_uri") {
			continue
		}

		client.RedirectURI = append(client.RedirectURI, l.URL)
	}

	u, err := url.Parse(id)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse id as url")
	}

	data := microformats.Parse(bytes.NewReader(resp.Body()), u)

	for _, item := range data.Items {
		if len(item.Type) == 0 && !strings.EqualFold(item.Type[0], HApp) &&
			!strings.EqualFold(item.Type[0], HXApp) {
			continue
		}

		populateProperties(item.Properties, client)
	}

	populateRels(data.Rels, client)

	return client, nil
}

func populateProperties(src map[string][]interface{}, dst *domain.Client) {
	for key, property := range src {
		if len(property) == 0 {
			continue
		}

		switch key {
		case KeyName:
			dst.Name = getString(property)
		case KeyLogo:
			for i := range property {
				switch val := property[i].(type) {
				case string:
					dst.Logo = val
				case map[string]string:
					dst.Logo = val[ValueValue]
				}
			}
		case KeyURL:
			dst.URL = getString(property)
		}
	}
}

func populateRels(src map[string][]string, dst *domain.Client) {
	for key, values := range src {
		if !strings.EqualFold(key, RelRedirectURI) {
			continue
		}

		for i := range values {
			dst.RedirectURI = append(dst.RedirectURI, values[i])
		}
	}
}

func getString(property []interface{}) string {
	for i := range property {
		val, _ := property[i].(string)

		return val
	}

	return ""
}
