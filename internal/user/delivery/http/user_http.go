package http

import (
	"encoding/json"
	"strings"

	"github.com/fasthttp/router"
	"github.com/lestrrat-go/jwx/jwa"
	http "github.com/valyala/fasthttp"

	"source.toby3d.me/toby3d/middleware"
	"source.toby3d.me/toby3d/auth/internal/common"
	"source.toby3d.me/toby3d/auth/internal/domain"
	"source.toby3d.me/toby3d/auth/internal/token"
)

type (
	UserInformationResponse struct {
		Name  string `json:"name,omitempty"`
		URL   string `json:"url,omitempty"`
		Photo string `json:"photo,omitempty"`
		Email string `json:"email,omitempty"`
	}

	RequestHandler struct {
		config *domain.Config
		tokens token.UseCase
	}
)

func NewRequestHandler(tokens token.UseCase, config *domain.Config) *RequestHandler {
	return &RequestHandler{
		tokens: tokens,
		config: config,
	}
}

func (h *RequestHandler) Register(r *router.Router) {
	chain := middleware.Chain{
		middleware.JWTWithConfig(middleware.JWTConfig{
			AuthScheme:    "Bearer",
			ContextKey:    "token",
			SigningKey:    []byte(h.config.JWT.Secret),
			SigningMethod: jwa.SignatureAlgorithm(h.config.JWT.Algorithm),
			Skipper:       middleware.DefaultSkipper,
			TokenLookup:   "header:" + http.HeaderAuthorization + ":Bearer ",
		}),
		middleware.LogFmt(),
	}

	r.GET("/userinfo", chain.RequestHandler(h.handleUserInformation))
}

func (h *RequestHandler) handleUserInformation(ctx *http.RequestCtx) {
	ctx.SetContentType(common.MIMEApplicationJSONCharsetUTF8)
	ctx.SetStatusCode(http.StatusOK)

	encoder := json.NewEncoder(ctx)

	tkn, userInfo, err := h.tokens.Verify(ctx, strings.TrimPrefix(string(ctx.Request.Header.Peek(
		http.HeaderAuthorization)), "Bearer "))
	if err != nil || tkn == nil {
		// WARN(toby3d): If the token is not valid, the endpoint still
		// MUST return a 200 Response.
		_ = encoder.Encode(err)

		return
	}

	if !tkn.Scope.Has(domain.ScopeProfile) {
		ctx.SetStatusCode(http.StatusForbidden)

		_ = encoder.Encode(domain.NewError(
			domain.ErrorCodeInsufficientScope,
			"token with 'profile' scope is required to view profile data",
			"https://indieauth.net/source/#user-information",
		))

		return
	}

	resp := new(UserInformationResponse)
	if userInfo == nil {
		_ = encoder.Encode(resp)

		return
	}

	if userInfo.HasName() {
		resp.Name = userInfo.GetName()
	}

	if userInfo.HasURL() {
		resp.URL = userInfo.GetURL().String()
	}

	if userInfo.HasPhoto() {
		resp.Photo = userInfo.GetPhoto().String()
	}

	if tkn.Scope.Has(domain.ScopeEmail) && userInfo.HasEmail() {
		resp.Email = userInfo.GetEmail().String()
	}

	_ = encoder.Encode(resp)
}
