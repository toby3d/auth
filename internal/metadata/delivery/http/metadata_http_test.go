package http_test

import (
	"testing"

	"github.com/fasthttp/router"
	"github.com/goccy/go-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	http "github.com/valyala/fasthttp"

	"source.toby3d.me/website/indieauth/internal/domain"
	delivery "source.toby3d.me/website/indieauth/internal/metadata/delivery/http"
	"source.toby3d.me/website/indieauth/internal/testing/httptest"
)

func TestMetadata(t *testing.T) {
	t.Parallel()

	r := router.New()
	cfg := domain.TestConfig(t)
	delivery.NewRequestHandler(cfg).Register(r)

	client, _, cleanup := httptest.New(t, r.Handler)
	t.Cleanup(cleanup)

	status, body, err := client.Get(nil, "https://example.com/.well-known/oauth-authorization-server")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, status)

	result := new(delivery.MetadataResponse)
	require.NoError(t, json.Unmarshal(body, result))
	assert.Equal(t, &delivery.MetadataResponse{
		AuthorizationEndpoint: cfg.Server.GetRootURL() + "authorize",
		Issuer:                cfg.Server.GetRootURL(),
		ServiceDocumentation:  "https://indieauth.net/source/",
		TokenEndpoint:         cfg.Server.GetRootURL() + "token",
		AuthorizationResponseIssParameterSupported: true,
		GrantTypesSupported: []string{
			domain.GrantTypeAuthorizationCode.String(),
		},
		ResponseTypesSupported: []string{
			domain.ResponseTypeCode.String(),
			domain.ResponseTypeID.String(),
		},
		CodeChallengeMethodsSupported: []string{
			domain.CodeChallengeMethodMD5.String(),
			domain.CodeChallengeMethodPLAIN.String(),
			domain.CodeChallengeMethodS1.String(),
			domain.CodeChallengeMethodS256.String(),
			domain.CodeChallengeMethodS512.String(),
		},
		ScopesSupported: []string{
			domain.ScopeEmail.String(),
			domain.ScopeProfile.String(),
		},
	}, result)
}
