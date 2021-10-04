package github

import (
	"context"

	json "github.com/goccy/go-json"
	"github.com/pkg/errors"
	http "github.com/valyala/fasthttp"
	"golang.org/x/oauth2"

	"source.toby3d.me/website/oauth/internal/domain"
	"source.toby3d.me/website/oauth/internal/profile"
)

type (
	//nolint: tagliatelle
	Response struct {
		Name      string `json:"name"`
		Blog      string `json:"blog"`
		AvatarURL string `json:"avatar_url"`
		Email     string `json:"email"`
	}

	githubProfileRepository struct {
		request *http.Request
		client  *http.Client
	}
)

func NewGitHubProfileRepository(client *http.Client) profile.Repository {
	req := http.AcquireRequest()
	req.SetRequestURI("https://api.github.com/user")
	req.Header.SetMethod(http.MethodGet)
	req.Header.Set(http.HeaderAccept, "application/vnd.github.v3+json")

	return &githubProfileRepository{
		request: req,
		client:  client,
	}
}

func (repo *githubProfileRepository) Get(ctx context.Context, token oauth2.Token) (*domain.Profile, error) {
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
		Name:  result.Name,
		URL:   result.Blog,
		Photo: result.AvatarURL,
		Email: result.Email,
	}, nil
}
