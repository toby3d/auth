package github_test

import (
	"context"
	"testing"
	"time"

	json "github.com/goccy/go-json"
	"github.com/stretchr/testify/assert"
	http "github.com/valyala/fasthttp"
	"golang.org/x/oauth2"

	"source.toby3d.me/website/oauth/internal/common"
	"source.toby3d.me/website/oauth/internal/domain"
	"source.toby3d.me/website/oauth/internal/profile/repository/github"
	"source.toby3d.me/website/oauth/internal/util"
)

func TestGet(t *testing.T) {
	t.Parallel()

	p := domain.TestProfile(t)
	client, _, cleanup := util.TestServe(t, func(ctx *http.RequestCtx) {
		ctx.SetStatusCode(http.StatusOK)
		ctx.SetContentType(common.MIMEApplicationJSON)
		_ = json.NewEncoder(ctx).Encode(&github.Response{
			Name:      p.Name,
			Blog:      p.URL,
			AvatarURL: p.Photo,
			Email:     p.Email,
		})
	})
	t.Cleanup(cleanup)

	result, err := github.NewGitHubProfileRepository(client).Get(context.TODO(), oauth2.Token{
		AccessToken:  "hackme",
		TokenType:    "Bearer",
		RefreshToken: "",
		Expiry:       time.Time{},
	})
	assert.NoError(t, err)
	assert.Equal(t, p, result)
}
