package http

import (
	"bytes"
	"context"
	"net/url"

	"github.com/pkg/errors"
	http "github.com/valyala/fasthttp"
	"willnorris.com/go/microformats"

	"source.toby3d.me/website/oauth/internal/authn"
)

type httpAuthnRepository struct {
	client *http.Client
}

func NewHTTPAuthnRepository(client *http.Client) authn.Repository {
	return &httpAuthnRepository{
		client: client,
	}
}

func (repo *httpAuthnRepository) Fetch(ctx context.Context, me string) ([]string, error) {
	u, err := url.Parse(me)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse me as url")
	}

	req := http.AcquireRequest()
	defer http.ReleaseRequest(req)
	req.SetRequestURI(u.String())
	req.Header.SetMethod(http.MethodGet)

	resp := http.AcquireResponse()
	defer http.ReleaseResponse(resp)

	if err := repo.client.Do(req, resp); err != nil {
		return nil, errors.Wrap(err, "failed to make a request to the entered me")
	}

	data := microformats.Parse(bytes.NewReader(resp.Body()), u)
	authn := make([]string, 0)

	for rel, values := range data.Rels {
		if rel != "authn" {
			continue
		}

		authn = append(authn, values...)
	}

	return authn, nil
}
