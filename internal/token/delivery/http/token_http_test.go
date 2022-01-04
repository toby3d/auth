package http_test

import (
	"context"
	"strings"
	"sync"
	"testing"

	"github.com/goccy/go-json"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	http "github.com/valyala/fasthttp"

	"source.toby3d.me/website/indieauth/internal/common"
	configrepo "source.toby3d.me/website/indieauth/internal/config/repository/viper"
	configucase "source.toby3d.me/website/indieauth/internal/config/usecase"
	"source.toby3d.me/website/indieauth/internal/domain"
	delivery "source.toby3d.me/website/indieauth/internal/token/delivery/http"
	repository "source.toby3d.me/website/indieauth/internal/token/repository/memory"
	"source.toby3d.me/website/indieauth/internal/token/usecase"
	"source.toby3d.me/website/indieauth/internal/util"
)

func TestVerification(t *testing.T) {
	t.Parallel()

	v := viper.New()
	v.SetDefault("indieauth.jwtSigningAlgorithm", "HS256")
	v.SetDefault("indieauth.jwtSecret", "hackme")

	accessToken := domain.TestToken(t)

	client, _, cleanup := util.TestServe(t, delivery.NewRequestHandler(usecase.NewTokenUseCase(
		repository.NewMemoryTokenRepository(new(sync.Map)),
		configucase.NewConfigUseCase(configrepo.NewViperConfigRepository(v)),
	)).Read)
	t.Cleanup(cleanup)

	req := http.AcquireRequest()
	defer http.ReleaseRequest(req)
	req.Header.SetMethod(http.MethodGet)
	req.SetRequestURI("https://app.example.com/token")
	req.Header.Set(http.HeaderAccept, common.MIMEApplicationJSON)
	req.Header.Set(http.HeaderAuthorization, "Bearer "+accessToken.AccessToken)

	resp := http.AcquireResponse()
	defer http.ReleaseResponse(resp)

	require.NoError(t, client.Do(req, resp))

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

	v := viper.New()
	v.SetDefault("indieauth.jwtSigningAlgorithm", "HS256")
	v.SetDefault("indieauth.jwtSecret", "hackme")

	tokens := repository.NewMemoryTokenRepository(new(sync.Map))
	accessToken := domain.TestToken(t)

	client, _, cleanup := util.TestServe(t, delivery.NewRequestHandler(
		usecase.NewTokenUseCase(tokens, configucase.NewConfigUseCase(configrepo.NewViperConfigRepository(v))),
	).Update)
	t.Cleanup(cleanup)

	req := http.AcquireRequest()
	defer http.ReleaseRequest(req)
	req.Header.SetMethod(http.MethodPost)
	req.SetRequestURI("https://app.example.com/token")
	req.Header.SetContentType(common.MIMEApplicationXWWWFormUrlencoded)
	req.Header.Set(http.HeaderAccept, common.MIMEApplicationJSON)
	req.PostArgs().Set("action", "revoke")
	req.PostArgs().Set("token", accessToken.AccessToken)

	resp := http.AcquireResponse()
	defer http.ReleaseResponse(resp)

	require.NoError(t, client.Do(req, resp))

	assert.Equal(t, http.StatusOK, resp.StatusCode())
	assert.Equal(t, `{}`, strings.TrimSpace(string(resp.Body())))

	result, err := tokens.Get(context.TODO(), accessToken.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, accessToken.AccessToken, result.AccessToken)
}
