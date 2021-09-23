package http

import (
	"bytes"
	"context"
	"net/url"
	"strings"

	"github.com/tomnomnom/linkheader"
	http "github.com/valyala/fasthttp"
	"source.toby3d.me/website/oauth/internal/client"
	"source.toby3d.me/website/oauth/internal/model"
	"willnorris.com/go/microformats"
)

type httpClientRepository struct {
	client *http.Client
}

func NewHTTPClientRepository(c *http.Client) client.Repository {
	return &httpClientRepository{
		client: c,
	}
}

func (repo *httpClientRepository) Get(ctx context.Context, id string) (*model.Client, error) {
	req := http.AcquireRequest()
	defer http.ReleaseRequest(req)

	req.Header.SetMethod(http.MethodGet)
	req.SetRequestURI(id)

	resp := http.AcquireResponse()
	defer http.ReleaseResponse(resp)

	if err := repo.client.Do(req, resp); err != nil {
		return nil, err
	}

	client := new(model.Client)
	client.ID = model.URL(id)
	client.RedirectURI = make([]model.URL, 0)

	for _, l := range linkheader.Parse(string(resp.Header.Peek(http.HeaderLink))) {
		if !strings.Contains(l.Rel, "redirect_uri") {
			continue
		}

		client.RedirectURI = append(client.RedirectURI, model.URL(l.URL))
	}

	u, err := url.Parse(id)
	if err != nil {
		return nil, err
	}

	data := microformats.Parse(bytes.NewReader(resp.Body()), u)

	populateItems(client, data.Items)
	populateRels(client, data.Rels)

	return client, nil
}

func populateItems(c *model.Client, items []*microformats.Microformat) {
	for _, item := range items {
		if len(item.Type) == 0 && item.Type[0] != "h-app" && item.Type[0] != "h-x-app" {
			continue
		}

		for key, property := range item.Properties {
			if len(property) == 0 {
				continue
			}

			switch key {
			case "name":
				for i := range property {
					val, _ := property[i].(string)
					c.Name = model.URL(val)
				}
			case "logo":
				for i := range property {
					switch val := property[i].(type) {
					case string:
						c.Logo = model.URL(val)
					case map[string]string:
						c.Logo = model.URL(val["value"])
					}
				}
			case "url":
				for i := range property {
					val, _ := property[i].(string)
					c.URL = model.URL(val)
				}
			}
		}
	}
}

func populateRels(c *model.Client, rels map[string][]string) {
	for key, values := range rels {
		if key != "redirect_uri" {
			continue
		}

		for i := range values {
			c.RedirectURI = append(c.RedirectURI, model.URL(values[i]))
		}
	}
}
