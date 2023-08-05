package http

import (
	"net/http"

	"github.com/goccy/go-json"

	"source.toby3d.me/toby3d/auth/internal/common"
	"source.toby3d.me/toby3d/auth/internal/domain"
)

type Handler struct {
	metadata *domain.Metadata
}

func NewHandler(metadata *domain.Metadata) *Handler {
	return &Handler{
		metadata: metadata,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "" && r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)

		return
	}

	w.Header().Set(common.HeaderContentType, common.MIMEApplicationJSONCharsetUTF8)

	scopes, responseTypes, grantTypes, codeChallengeMethods := make([]string, 0), make([]string, 0),
		make([]string, 0), make([]string, 0)

	for i := range h.metadata.ScopesSupported {
		scopes = append(scopes, h.metadata.ScopesSupported[i].String())
	}

	for i := range h.metadata.ResponseTypesSupported {
		responseTypes = append(responseTypes, h.metadata.ResponseTypesSupported[i].String())
	}

	for i := range h.metadata.GrantTypesSupported {
		grantTypes = append(grantTypes, h.metadata.GrantTypesSupported[i].String())
	}

	for i := range h.metadata.CodeChallengeMethodsSupported {
		codeChallengeMethods = append(codeChallengeMethods,
			h.metadata.CodeChallengeMethodsSupported[i].String())
	}

	_ = json.NewEncoder(w).Encode(&MetadataResponse{
		AuthorizationEndpoint: h.metadata.AuthorizationEndpoint.String(),
		IntrospectionEndpoint: h.metadata.IntrospectionEndpoint.String(),
		Issuer:                h.metadata.Issuer.String(),
		RevocationEndpoint:    h.metadata.RevocationEndpoint.String(),
		ServiceDocumentation:  h.metadata.ServiceDocumentation.String(),
		TokenEndpoint:         h.metadata.TokenEndpoint.String(),
		UserinfoEndpoint:      h.metadata.UserinfoEndpoint.String(),
		AuthorizationResponseIssParameterSupported: h.metadata.AuthorizationResponseIssParameterSupported,
		CodeChallengeMethodsSupported:              codeChallengeMethods,
		GrantTypesSupported:                        grantTypes,
		IntrospectionEndpointAuthMethodsSupported:  h.metadata.IntrospectionEndpointAuthMethodsSupported,
		ResponseTypesSupported:                     responseTypes,
		ScopesSupported:                            scopes,
		// NOTE(toby3d): If a revocation endpoint is provided, this
		// property should also be provided with the value ["none"],
		// since the omission of this value defaults to
		// client_secret_basic according to RFC8414.
		RevocationEndpointAuthMethodsSupported: h.metadata.RevocationEndpointAuthMethodsSupported,
	})

	w.WriteHeader(http.StatusOK)
}
