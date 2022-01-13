package http

import (
	"fmt"
	"path"
	"strings"

	"github.com/fasthttp/router"
	json "github.com/goccy/go-json"
	http "github.com/valyala/fasthttp"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/xerrors"

	"source.toby3d.me/toby3d/form"
	"source.toby3d.me/toby3d/middleware"
	"source.toby3d.me/website/indieauth/internal/auth"
	"source.toby3d.me/website/indieauth/internal/client"
	"source.toby3d.me/website/indieauth/internal/common"
	"source.toby3d.me/website/indieauth/internal/domain"
	"source.toby3d.me/website/indieauth/web"
)

type (
	AuthorizeRequest struct {
		// Indicates to the authorization server that an authorization
		// code should be returned as the response.
		ResponseType domain.ResponseType `form:"response_type"` // code

		// The client URL.
		ClientID *domain.ClientID `form:"client_id"`

		// The redirect URL indicating where the user should be
		// redirected to after approving the request.
		RedirectURI *domain.URL `form:"redirect_uri"`

		// A parameter set by the client which will be included when the
		// user is redirected back to the client. This is used to
		// prevent CSRF attacks. The authorization server MUST return
		// the unmodified state value back to the client.
		State string `form:"state"`

		// The code challenge as previously described.
		CodeChallenge string `form:"code_challenge"`

		// The hashing method used to calculate the code challenge.
		CodeChallengeMethod domain.CodeChallengeMethod `form:"code_challenge_method"`

		// A space-separated list of scopes the client is requesting,
		// e.g. "profile", or "profile create". If the client omits this
		// value, the authorization server MUST NOT issue an access
		// token for this authorization code. Only the user's profile
		// URL may be returned without any scope requested.
		Scope domain.Scopes `form:"scope"`

		// The URL that the user entered.
		Me *domain.Me `form:"me"`
	}

	VerifyRequest struct {
		ClientID            *domain.ClientID           `form:"client_id"`
		Me                  *domain.Me                 `form:"me"`
		RedirectURI         *domain.URL                `form:"redirect_uri"`
		CodeChallengeMethod domain.CodeChallengeMethod `form:"code_challenge_method"`
		ResponseType        domain.ResponseType        `form:"response_type"`
		Scope               domain.Scopes              `form:"scope[]"` // TODO(toby3d): fix parsing in form pkg
		Authorize           string                     `form:"authorize"`
		CodeChallenge       string                     `form:"code_challenge"`
		State               string                     `form:"state"`
	}

	ExchangeRequest struct {
		GrantType domain.GrantType `form:"grant_type"` // authorization_code

		// The authorization code received from the authorization
		// endpoint in the redirect.
		Code string `form:"code"`

		// The client's URL, which MUST match the client_id used in the
		// authentication request.
		ClientID *domain.ClientID `form:"client_id"`

		// The client's redirect URL, which MUST match the initial
		// authentication request.
		RedirectURI *domain.URL `form:"redirect_uri"`

		// The original plaintext random string generated before
		// starting the authorization request.
		CodeVerifier string `form:"code_verifier"`
	}

	ExchangeResponse struct {
		Me *domain.Me `json:"me"`
	}

	NewRequestHandlerOptions struct {
		Auth    auth.UseCase
		Clients client.UseCase
		Config  *domain.Config
		Matcher language.Matcher
	}

	RequestHandler struct {
		clients client.UseCase
		config  *domain.Config
		matcher language.Matcher
		useCase auth.UseCase
	}
)

func NewRequestHandler(opts NewRequestHandlerOptions) *RequestHandler {
	return &RequestHandler{
		clients: opts.Clients,
		config:  opts.Config,
		matcher: opts.Matcher,
		useCase: opts.Auth,
	}
}

func (h *RequestHandler) Register(r *router.Router) {
	chain := middleware.Chain{
		middleware.CSRFWithConfig(middleware.CSRFConfig{
			Skipper: func(ctx *http.RequestCtx) bool {
				matched, _ := path.Match("/api/*", string(ctx.Path()))

				return ctx.IsPost() && matched
			},
			CookieSameSite: http.CookieSameSiteLaxMode,
			CookieName:     "_csrf",
			TokenLookup:    "form:_csrf",
			CookieSecure:   true,
			CookieHTTPOnly: true,
		}),
		middleware.LogFmt(),
	}

	r.GET("/authorize", chain.RequestHandler(h.handleRender))
	r.POST("/api/authorize", chain.RequestHandler(h.handleVerify))
	r.POST("/authorize", chain.RequestHandler(h.handleExchange))
}

func (h *RequestHandler) handleRender(ctx *http.RequestCtx) {
	req := new(AuthorizeRequest)
	if err := req.bind(ctx); err != nil {
		ctx.Error(err.Error(), http.StatusBadRequest)

		return
	}

	client, err := h.clients.Discovery(ctx, req.ClientID)
	if err != nil {
		ctx.Error(err.Error(), http.StatusBadRequest)

		return
	}

	if !client.ValidateRedirectURI(req.RedirectURI) {
		ctx.Error("requested redirect_uri is not registered on client_id side", http.StatusBadRequest)

		return
	}

	csrf, _ := ctx.UserValue(middleware.DefaultCSRFConfig.ContextKey).([]byte)
	tags, _, _ := language.ParseAcceptLanguage(string(ctx.Request.Header.Peek(http.HeaderAcceptLanguage)))
	tag, _, _ := h.matcher.Match(tags...)

	ctx.SetContentType(common.MIMETextHTMLCharsetUTF8)
	web.WriteTemplate(ctx, &web.AuthorizePage{
		BaseOf: web.BaseOf{
			Config:   h.config,
			Language: tag,
			Printer:  message.NewPrinter(tag),
		},
		Client:              client,
		CodeChallenge:       req.CodeChallenge,
		CodeChallengeMethod: req.CodeChallengeMethod,
		CSRF:                csrf,
		Me:                  req.Me,
		RedirectURI:         req.RedirectURI,
		ResponseType:        req.ResponseType,
		Scope:               req.Scope,
		State:               req.State,
	})
}

func (h *RequestHandler) handleVerify(ctx *http.RequestCtx) {
	ctx.SetContentType(common.MIMEApplicationJSONCharsetUTF8)

	encoder := json.NewEncoder(ctx)

	req := new(VerifyRequest)
	if err := req.bind(ctx); err != nil {
		ctx.SetStatusCode(http.StatusBadRequest)
		encoder.Encode(domain.Error{
			Code:        "invalid_request",
			Description: err.Error(),
			Frame:       xerrors.Caller(1),
		})

		return
	}

	u := http.AcquireURI()
	defer http.ReleaseURI(u)
	req.RedirectURI.CopyTo(u)

	if strings.EqualFold(req.Authorize, "deny") {
		u.QueryArgs().Set("error", "access_denied")
		u.QueryArgs().Set("error_description", "user deny authorization request")
		ctx.Redirect(u.String(), http.StatusFound)

		return
	}

	code, err := h.useCase.Generate(ctx, auth.GenerateOptions{
		ClientID:            req.ClientID,
		RedirectURI:         req.RedirectURI,
		CodeChallenge:       req.CodeChallenge,
		CodeChallengeMethod: req.CodeChallengeMethod,
		Scope:               req.Scope,
		Me:                  req.Me,
	})
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		encoder.Encode(domain.Error{
			Description: err.Error(),
			Frame:       xerrors.Caller(1),
		})

		return
	}

	for key, val := range map[string]string{
		"code":  code,
		"iss":   h.config.Server.GetRootURL(),
		"state": req.State,
	} {
		u.QueryArgs().Set(key, val)
	}

	ctx.Redirect(u.String(), http.StatusFound)
}

func (h *RequestHandler) handleExchange(ctx *http.RequestCtx) {
	ctx.SetContentType(common.MIMEApplicationJSONCharsetUTF8)

	encoder := json.NewEncoder(ctx)

	req := new(ExchangeRequest)
	if err := req.bind(ctx); err != nil {
		ctx.SetStatusCode(http.StatusBadRequest)
		encoder.Encode(err)

		return
	}

	me, err := h.useCase.Exchange(ctx, auth.ExchangeOptions{
		Code:         req.Code,
		ClientID:     req.ClientID,
		RedirectURI:  req.RedirectURI,
		CodeVerifier: req.CodeVerifier,
	})
	if err != nil {
		ctx.SetStatusCode(http.StatusBadRequest)
		encoder.Encode(err)

		return
	}

	encoder.Encode(&ExchangeResponse{
		Me: me,
	})
}

func (r *AuthorizeRequest) bind(ctx *http.RequestCtx) error {
	if err := form.Unmarshal(ctx.QueryArgs(), r); err != nil {
		return domain.Error{
			Code:        "invalid_request",
			Description: err.Error(),
			Frame:       xerrors.Caller(1),
		}
	}

	r.Scope = make(domain.Scopes, 0)
	parseScope(r.Scope, ctx.QueryArgs().Peek("scope"))

	if r.ResponseType == domain.ResponseTypeID {
		r.ResponseType = domain.ResponseTypeCode
	}

	return nil
}

func (r *VerifyRequest) bind(ctx *http.RequestCtx) error {
	if err := form.Unmarshal(ctx.PostArgs(), r); err != nil {
		return domain.Error{
			Code:        "invalid_request",
			Description: err.Error(),
			Frame:       xerrors.Caller(1),
		}
	}

	r.Scope = make(domain.Scopes, 0)
	parseScope(r.Scope, ctx.PostArgs().PeekMulti("scope[]")...)

	if r.ResponseType == domain.ResponseTypeID {
		r.ResponseType = domain.ResponseTypeCode
	}

	if !strings.EqualFold(r.Authorize, "allow") && !strings.EqualFold(r.Authorize, "deny") {
		return domain.Error{
			Code:        "invalid_request",
			Description: "cannot validate verification request",
			Frame:       xerrors.Caller(1),
		}
	}

	return nil
}

func (r *ExchangeRequest) bind(ctx *http.RequestCtx) error {
	if err := form.Unmarshal(ctx.PostArgs(), r); err != nil {
		return domain.Error{
			Code:        "invalid_request",
			Description: err.Error(),
			Frame:       xerrors.Caller(1),
		}
	}

	return nil
}

// TODO(toby3d): fix this in form pkg.
func parseScope(dst domain.Scopes, src ...[]byte) error {
	if len(src) == 0 {
		return nil
	}

	var scopes []string

	if len(src) == 1 {
		scopes = strings.Fields(string(src[0]))
	}

	for _, rawScope := range scopes {
		scope, err := domain.ParseScope(string(rawScope))
		if err != nil {
			return &domain.Error{
				Code:        "invalid_request",
				Description: fmt.Sprintf("cannot parse scope: %v", err),
				Frame:       xerrors.Caller(1),
			}
		}

		dst = append(dst, scope)
	}

	return nil
}
