package mastodon_test

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
	"source.toby3d.me/website/oauth/internal/profile/repository/mastodon"
	"source.toby3d.me/website/oauth/internal/util"
)

func TestGet(t *testing.T) {
	t.Parallel()

	p := domain.TestProfile(t)
	p.Email = "" // WARN(toby3d): Mastodon does not provide user email information

	client, _, cleanup := util.TestServe(t, func(ctx *http.RequestCtx) {
		ctx.SetStatusCode(http.StatusOK)
		ctx.SetContentType(common.MIMEApplicationJSON)
		_ = json.NewEncoder(ctx).Encode(&mastodon.Response{
			DisplayName: p.Name,
			Avatar:      p.Photo,
			URL:         p.URL,
		})
	})
	t.Cleanup(cleanup)

	result, err := mastodon.NewMastodonProfileRepository(client, "https://mstdn.io/").
		Get(context.TODO(), oauth2.Token{
			AccessToken:  "hackme",
			TokenType:    "Bearer",
			RefreshToken: "",
			Expiry:       time.Time{},
		})
	assert.NoError(t, err)
	assert.Equal(t, p, result)
}
