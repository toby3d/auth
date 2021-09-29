package http_test

import (
	"context"
	"strings"
	"sync"
	"testing"

	"github.com/goccy/go-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	http "github.com/valyala/fasthttp"

	"source.toby3d.me/website/oauth/internal/common"
	"source.toby3d.me/website/oauth/internal/domain"
	delivery "source.toby3d.me/website/oauth/internal/token/delivery/http"
	repository "source.toby3d.me/website/oauth/internal/token/repository/memory"
	"source.toby3d.me/website/oauth/internal/token/usecase"
	"source.toby3d.me/website/oauth/internal/util"
)

func TestVerification(t *testing.T) {
	t.Parallel()

	store := new(sync.Map)
	repo := repository.NewMemoryTokenRepository(store)
	accessToken := domain.TestToken(t)

	require.NoError(t, repo.Create(context.TODO(), accessToken))

	req := http.AcquireRequest()
	defer http.ReleaseRequest(req)

	req.Header.SetMethod(http.MethodGet)
	req.SetRequestURI("http://localhost/token")
	req.Header.Set(http.HeaderAccept, common.MIMEApplicationJSON)
	req.Header.Set(http.HeaderAuthorization, "Bearer "+accessToken.AccessToken)

	resp := http.AcquireResponse()
	defer http.ReleaseResponse(resp)

	require.NoError(t, util.Serve(delivery.NewRequestHandler(usecase.NewTokenUseCase(repo)).Read, req, resp))
	assert.Equal(t, http.StatusOK, resp.StatusCode())

	token := new(delivery.VerificationResponse)
	require.NoError(t, json.Unmarshal(resp.Body(), token))
	assert.Equal(t, &delivery.VerificationResponse{
		Me:       accessToken.Me,
		ClientID: accessToken.ClientID,
		Scope:    strings.Join(accessToken.Scopes, " "),
	}, token)
}

func TestRevocation(t *testing.T) {
	t.Parallel()

	store := new(sync.Map)
	repo := repository.NewMemoryTokenRepository(store)
	accessToken := domain.TestToken(t)

	require.NoError(t, repo.Create(context.TODO(), domain.TestToken(t)))

	req := http.AcquireRequest()
	defer http.ReleaseRequest(req)

	req.Header.SetMethod(http.MethodPost)
	req.SetRequestURI("http://localhost/token")
	req.Header.SetContentType(common.MIMEApplicationXWWWFormUrlencoded)
	req.Header.Set(http.HeaderAccept, common.MIMEApplicationJSON)
	req.PostArgs().Set("action", "revoke")
	req.PostArgs().Set("token", accessToken.AccessToken)

	resp := http.AcquireResponse()
	defer http.ReleaseResponse(resp)

	require.NoError(t, util.Serve(delivery.NewRequestHandler(usecase.NewTokenUseCase(repo)).Update, req, resp))
	assert.Equal(t, http.StatusOK, resp.StatusCode())
	assert.Equal(t, `{}`, strings.TrimSpace(string(resp.Body())))

	token, err := repo.Get(context.TODO(), accessToken.AccessToken)
	require.NoError(t, err)
	assert.Nil(t, token)
}
