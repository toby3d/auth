package http

import (
	"encoding/json"
	"strings"

	"github.com/fasthttp/router"
	"github.com/lestrrat-go/jwx/jwa"
	http "github.com/valyala/fasthttp"

	"source.toby3d.me/toby3d/middleware"
	"source.toby3d.me/website/indieauth/internal/common"
	"source.toby3d.me/website/indieauth/internal/domain"
	"source.toby3d.me/website/indieauth/internal/token"
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
			AuthScheme:              "Bearer",
			BeforeFunc:              nil,
			Claims:                  nil,
			ContextKey:              "token",
			ErrorHandler:            nil,
			ErrorHandlerWithContext: nil,
			ParseTokenFunc:          nil,
			SigningKey:              []byte(h.config.JWT.Secret),
			SigningKeys:             nil,
			SigningMethod:           jwa.SignatureAlgorithm(h.config.JWT.Algorithm),
			Skipper:                 middleware.DefaultSkipper,
			SuccessHandler:          nil,
			TokenLookup: middleware.SourceHeader + ":" + http.HeaderAuthorization +
				"," + middleware.SourceCookie + ":" + "__Secure-auth-token",
		}),
		middleware.LogFmt(),
	}

	r.GET("/userinfo", chain.RequestHandler(h.handleUserInformation))
}

func (h *RequestHandler) handleUserInformation(ctx *http.RequestCtx) {
	ctx.SetContentType(common.MIMEApplicationJSONCharsetUTF8)
	ctx.SetStatusCode(http.StatusOK)

	encoder := json.NewEncoder(ctx)

	tkn, err := h.tokens.Verify(ctx, strings.TrimPrefix(string(ctx.Request.Header.Peek(http.HeaderAuthorization)),
		"Bearer "))
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
	if tkn.Profile == nil {
		_ = encoder.Encode(resp)

		return
	}

	resp.Name = tkn.Profile.GetName()

	if url := tkn.Profile.GetURL(); url != nil {
		resp.URL = url.String()
	}

	if photo := tkn.Profile.GetPhoto(); photo != nil {
		resp.Photo = photo.String()
	}

	if email := tkn.Profile.GetEmail(); email != nil && tkn.Scope.Has(domain.ScopeEmail) {
		resp.Email = email.String()
	}

	_ = encoder.Encode(resp)
}
