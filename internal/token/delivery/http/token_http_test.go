package http_test

import (
	"context"
	"strings"
	"sync"
	"testing"

	"github.com/brianvoe/gofakeit"
	"github.com/goccy/go-json"
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

func TestVerification(t *testing.T) {
	t.Parallel()

	require := require.New(t)
	assert := assert.New(t)
	store := new(sync.Map)
	repo := repository.NewMemoryTokenRepository(store)
	accessToken := model.Token{
		AccessToken: gofakeit.Password(true, true, true, true, false, 32),
		Type:        "Bearer",
		ClientID:    "https://app.example.com/",
		Me:          "https://user.example.net/",
		Scopes:      []string{"create", "update", "delete"},
		Profile:     nil,
	}

	require.NoError(repo.Create(context.TODO(), &accessToken))

	req := http.AcquireRequest()
	defer http.ReleaseRequest(req)

	req.Header.SetMethod(http.MethodGet)
	req.SetRequestURI("http://localhost/token")
	req.Header.Set(http.HeaderAccept, common.MIMEApplicationJSON)
	req.Header.Set(http.HeaderAuthorization, "Bearer "+accessToken.AccessToken)

	resp := http.AcquireResponse()
	defer http.ReleaseResponse(resp)

	require.NoError(util.Serve(delivery.NewRequestHandler(usecase.NewTokenUseCase(repo)).Read, req, resp))
	assert.Equal(http.StatusOK, resp.StatusCode())

	token := new(delivery.VerificationResponse)
	require.NoError(json.Unmarshal(resp.Body(), token))
	assert.Equal(&delivery.VerificationResponse{
		Me:       accessToken.Me,
		ClientID: accessToken.ClientID,
		Scope:    strings.Join(accessToken.Scopes, " "),
	}, token)
}

func TestRevocation(t *testing.T) {
	t.Parallel()

	require := require.New(t)
	assert := assert.New(t)
	store := new(sync.Map)
	repo := repository.NewMemoryTokenRepository(store)
	accessToken := gofakeit.Password(true, true, true, true, false, 32)

	require.NoError(repo.Create(context.TODO(), &model.Token{
		AccessToken: accessToken,
		Type:        "Bearer",
		ClientID:    "https://app.example.com/",
		Me:          "https://user.example.net/",
		Scopes:      []string{"create", "update", "delete"},
		Profile:     nil,
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

	require.NoError(util.Serve(delivery.NewRequestHandler(usecase.NewTokenUseCase(repo)).Update, req, resp))
	assert.Equal(http.StatusOK, resp.StatusCode())
	assert.Equal(`{}`, strings.TrimSpace(string(resp.Body())))

	token, err := repo.Get(context.TODO(), accessToken)
	require.NoError(err)
	assert.Nil(token)
}
