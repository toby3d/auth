package mastodon

import (
	"context"
	"path"

	json "github.com/goccy/go-json"
	"github.com/pkg/errors"
	http "github.com/valyala/fasthttp"
	"golang.org/x/oauth2"

	"source.toby3d.me/website/oauth/internal/common"
	"source.toby3d.me/website/oauth/internal/domain"
	"source.toby3d.me/website/oauth/internal/profile"
)

type (
	//nolint: tagliatelle
	Response struct {
		DisplayName string `json:"display_name"`
		Avatar      string `json:"avatar"`
		URL         string `json:"url"`
	}

	mastodonProfileRepository struct {
		request *http.Request
		client  *http.Client
	}
)

func NewMastodonProfileRepository(client *http.Client, baseURL string) profile.Repository {
	req := http.AcquireRequest()
	req.SetRequestURI(baseURL)
	req.URI().SetPath(path.Join("api", "v1", "accounts", "verify_credentials"))
	req.Header.SetMethod(http.MethodGet)
	req.Header.Set(http.HeaderAccept, common.MIMEApplicationJSON)

	return &mastodonProfileRepository{
		request: req,
		client:  client,
	}
}

func (repo *mastodonProfileRepository) Get(ctx context.Context, token oauth2.Token) (*domain.Profile, error) {
	req := http.AcquireRequest()
	defer http.ReleaseRequest(req)
	repo.request.CopyTo(req)
	req.Header.Set(http.HeaderAuthorization, token.TokenType+" "+token.AccessToken)

	resp := http.AcquireResponse()
	defer http.ReleaseResponse(resp)

	if err := repo.client.Do(req, resp); err != nil {
		return nil, errors.Wrap(err, "failed to fetch authenticated user")
	}

	result := new(Response)
	if err := json.Unmarshal(resp.Body(), result); err != nil {
		return nil, errors.Wrap(err, "cannot unmarshal GitHub response")
	}

	return &domain.Profile{
		Name:  result.DisplayName,
		URL:   result.URL,
		Photo: result.Avatar,
		Email: "",
	}, nil
}