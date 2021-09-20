package http_test

import (
	"context"
	"strings"
	"testing"

	"github.com/brianvoe/gofakeit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	http "github.com/valyala/fasthttp"

	"source.toby3d.me/website/oauth/internal/common"
	"source.toby3d.me/website/oauth/internal/model"
	delivery "source.toby3d.me/website/oauth/internal/token/delivery/http"
	repository "source.toby3d.me/website/oauth/internal/token/repository/memory"
	"source.toby3d.me/website/oauth/internal/token/usecase"
	"source.toby3d.me/website/oauth/internal/util"
)

func TestRevocation(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)
	repo := repository.NewMemoryTokenRepository()
	accessToken := gofakeit.Password(true, true, true, true, false, 32)

	require.NoError(repo.Create(context.TODO(), &model.Token{
		AccessToken: accessToken,
		Type:        "Bearer",
		ClientID:    "https://app.example.com/",
		Me:          "https://user.example.net/",
		Scopes:      []string{"create", "update", "delete"},
	}))

	req := http.AcquireRequest()
	defer http.ReleaseRequest(req)

	req.Header.SetMethod(http.MethodPost)
	req.SetRequestURI("http://localhost/token")
	req.Header.SetContentType(common.MIMEApplicationXWWWFormUrlencoded)
	req.Header.Set(http.HeaderAccept, common.MIMEApplicationJSON)
	req.PostArgs().Set("action", "revoke")
	req.PostArgs().Set("token", accessToken)

	resp := http.AcquireResponse()
	defer http.ReleaseResponse(resp)

	require.NoError(util.Serve(delivery.NewRequestHandler(usecase.NewTokenUseCase(repo)).Revocation, req, resp))
	assert.Equal(http.StatusOK, resp.StatusCode())
	assert.Equal(`{}`, strings.TrimSpace(string(resp.Body())))

	token, err := repo.Get(context.TODO(), accessToken)
	require.NoError(err)
	assert.Nil(token)
}
