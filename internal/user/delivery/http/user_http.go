package http

import (
	"net/http"
	"strings"

	"github.com/goccy/go-json"
	"github.com/lestrrat-go/jwx/v2/jwa"

	"source.toby3d.me/toby3d/auth/internal/common"
	"source.toby3d.me/toby3d/auth/internal/domain"
	"source.toby3d.me/toby3d/auth/internal/middleware"
	"source.toby3d.me/toby3d/auth/internal/token"
)

type Handler struct {
	config *domain.Config
	tokens token.UseCase
}

func NewHandler(tokens token.UseCase, config *domain.Config) *Handler {
	return &Handler{
		tokens: tokens,
		config: config,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	chain := middleware.Chain{
		//nolint:exhaustivestruct
		middleware.JWTWithConfig(middleware.JWTConfig{
			AuthScheme:    "Bearer",
			ContextKey:    "token",
			SigningKey:    []byte(h.config.JWT.Secret),
			SigningMethod: jwa.SignatureAlgorithm(h.config.JWT.Algorithm),
			Skipper:       middleware.DefaultSkipper,
			TokenLookup:   "header:" + common.HeaderAuthorization + ":Bearer ",
		}),
	}

	chain.Handler(h.handleFunc).ServeHTTP(w, r)
}

func (h *Handler) handleFunc(w http.ResponseWriter, r *http.Request) {
	if r.Method != "" && r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)

		return
	}

	w.Header().Set(common.HeaderContentType, common.MIMEApplicationJSONCharsetUTF8)

	encoder := json.NewEncoder(w)

	tkn, userInfo, err := h.tokens.Verify(r.Context(),
		strings.TrimPrefix(r.Header.Get(common.HeaderAuthorization), "Bearer "))
	if err != nil || tkn == nil {
		// WARN(toby3d): If the token is not valid, the endpoint still
		// MUST return a 200 Response.
		_ = encoder.Encode(err) //nolint:errchkjson

		w.WriteHeader(http.StatusOK)

		return
	}

	if !tkn.Scope.Has(domain.ScopeProfile) {
		//nolint:errchkjson
		_ = encoder.Encode(domain.NewError(
			domain.ErrorCodeInsufficientScope,
			"token with 'profile' scope is required to view profile data",
			"https://indieauth.net/source/#user-information",
		))

		w.WriteHeader(http.StatusForbidden)

		return
	}

	//nolint:errchkjson
	_ = encoder.Encode(NewUserInformationResponse(userInfo,
		userInfo.HasEmail() && tkn.Scope.Has(domain.ScopeEmail)))

	w.WriteHeader(http.StatusOK)
}
