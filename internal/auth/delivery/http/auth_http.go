package http

import (
	"bytes"
	"crypto/subtle"
	"errors"
	"path"
	"strings"

	"github.com/fasthttp/router"
	json "github.com/goccy/go-json"
	http "github.com/valyala/fasthttp"
	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"source.toby3d.me/toby3d/form"
	"source.toby3d.me/toby3d/middleware"
	"source.toby3d.me/website/indieauth/internal/auth"
	"source.toby3d.me/website/indieauth/internal/client"
	"source.toby3d.me/website/indieauth/internal/common"
	"source.toby3d.me/website/indieauth/internal/domain"
	"source.toby3d.me/website/indieauth/web"
)

type (
	AuthAuthorizeRequest struct {
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

	AuthVerifyRequest struct {
		ClientID            *domain.ClientID           `form:"client_id"`
		Me                  *domain.Me                 `form:"me"`
		RedirectURI         *domain.URL                `form:"redirect_uri"`
		CodeChallengeMethod domain.CodeChallengeMethod `form:"code_challenge_method"`
		ResponseType        domain.ResponseType        `form:"response_type"`
		Scope               domain.Scopes              `form:"scope[]"` // TODO(toby3d): fix parsing in form pkg
		Authorize           string                     `form:"authorize"`
		CodeChallenge       string                     `form:"code_challenge"`
		State               string                     `form:"state"`
		Provider            string                     `form:"provider"`
	}

	AuthExchangeRequest struct {
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

	AuthExchangeResponse struct {
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
				return ctx.IsPost() && string(ctx.Path()) == "/authorize"
			},
			CookieMaxAge:   0,
			CookieSameSite: http.CookieSameSiteStrictMode,
			ContextKey:     "",
			CookieDomain:   "",
			CookieName:     "_csrf",
			CookiePath:     "",
			TokenLookup:    "form:_csrf",
			TokenLength:    0,
			CookieSecure:   true,
			CookieHTTPOnly: true,
		}),
		middleware.BasicAuthWithConfig(middleware.BasicAuthConfig{
			Skipper: func(ctx *http.RequestCtx) bool {
				matched, _ := path.Match("/api/*", string(ctx.Path()))
				provider := string(ctx.QueryArgs().Peek("provider"))
				providerMatched := provider != "" && provider != domain.ProviderDirect.UID

				return !ctx.IsPost() || !matched || providerMatched
			},
			Validator: func(ctx *http.RequestCtx, login, password string) (bool, error) {
				userMatch := subtle.ConstantTimeCompare([]byte(login),
					[]byte(h.config.IndieAuth.Username))
				passMatch := subtle.ConstantTimeCompare([]byte(password),
					[]byte(h.config.IndieAuth.Password))

				return userMatch == 1 && passMatch == 1, nil
			},
			Realm: "",
		}),
		middleware.LogFmt(),
	}

	r.GET("/authorize", chain.RequestHandler(h.handleRender))
	r.POST("/api/authorize", chain.RequestHandler(h.handleVerify))
	r.POST("/authorize", chain.RequestHandler(h.handleExchange))
}

func (h *RequestHandler) handleRender(ctx *http.RequestCtx) {
	ctx.SetContentType(common.MIMETextHTMLCharsetUTF8)

	tags, _, _ := language.ParseAcceptLanguage(string(ctx.Request.Header.Peek(http.HeaderAcceptLanguage)))
	tag, _, _ := h.matcher.Match(tags...)
	baseOf := web.BaseOf{
		Config:   h.config,
		Language: tag,
		Printer:  message.NewPrinter(tag),
	}

	req := NewAuthAuthorizeRequest()
	if err := req.bind(ctx); err != nil {
		ctx.SetStatusCode(http.StatusBadRequest)
		web.WriteTemplate(ctx, &web.ErrorPage{
			BaseOf: baseOf,
			Error:  err,
		})

		return
	}

	client, err := h.clients.Discovery(ctx, req.ClientID)
	if err != nil {
		ctx.SetStatusCode(http.StatusBadRequest)
		web.WriteTemplate(ctx, &web.ErrorPage{
			BaseOf: baseOf,
			Error:  err,
		})

		return
	}

	if !client.ValidateRedirectURI(req.RedirectURI) {
		ctx.SetStatusCode(http.StatusBadRequest)
		web.WriteTemplate(ctx, &web.ErrorPage{
			BaseOf: baseOf,
			Error: domain.NewError(
				domain.ErrorCodeInvalidClient,
				"requested redirect_uri is not registered on client_id side",
				"",
			),
		})

		return
	}

	csrf, _ := ctx.UserValue(middleware.DefaultCSRFConfig.ContextKey).([]byte)
	web.WriteTemplate(ctx, &web.AuthorizePage{
		BaseOf:              baseOf,
		CSRF:                csrf,
		Scope:               req.Scope,
		Client:              client,
		Me:                  req.Me,
		RedirectURI:         req.RedirectURI,
		CodeChallengeMethod: req.CodeChallengeMethod,
		ResponseType:        req.ResponseType,
		CodeChallenge:       req.CodeChallenge,
		State:               req.State,
		Providers:           make([]*domain.Provider, 0), // TODO(toby3d)
	})
}

func (h *RequestHandler) handleVerify(ctx *http.RequestCtx) {
	ctx.SetContentType(common.MIMEApplicationJSONCharsetUTF8)

	encoder := json.NewEncoder(ctx)

	req := NewAuthVerifyRequest()
	if err := req.bind(ctx); err != nil {
		ctx.SetStatusCode(http.StatusBadRequest)

		_ = encoder.Encode(err)

		return
	}

	redirectURL := http.AcquireURI()
	defer http.ReleaseURI(redirectURL)
	req.RedirectURI.CopyTo(redirectURL)

	if strings.EqualFold(req.Authorize, "deny") {
		domain.NewError(domain.ErrorCodeAccessDenied, "user deny authorization request", "", req.State).
			SetReirectURI(redirectURL)
		ctx.Redirect(redirectURL.String(), http.StatusFound)

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

		_ = encoder.Encode(err)

		return
	}

	for key, val := range map[string]string{
		"code":  code,
		"iss":   h.config.Server.GetRootURL(),
		"state": req.State,
	} {
		redirectURL.QueryArgs().Set(key, val)
	}

	ctx.Redirect(redirectURL.String(), http.StatusFound)
}

func (h *RequestHandler) handleExchange(ctx *http.RequestCtx) {
	ctx.SetContentType(common.MIMEApplicationJSONCharsetUTF8)

	encoder := json.NewEncoder(ctx)

	req := new(AuthExchangeRequest)
	if err := req.bind(ctx); err != nil {
		ctx.SetStatusCode(http.StatusBadRequest)

		_ = encoder.Encode(err)

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

		_ = encoder.Encode(err)

		return
	}

	_ = encoder.Encode(&AuthExchangeResponse{
		Me: me,
	})
}

func NewAuthAuthorizeRequest() *AuthAuthorizeRequest {
	return &AuthAuthorizeRequest{
		ClientID:            new(domain.ClientID),
		CodeChallenge:       "",
		CodeChallengeMethod: domain.CodeChallengeMethodUndefined,
		Me:                  new(domain.Me),
		RedirectURI:         new(domain.URL),
		ResponseType:        domain.ResponseTypeUndefined,
		Scope:               make(domain.Scopes, 0),
		State:               "",
	}
}

func (r *AuthAuthorizeRequest) bind(ctx *http.RequestCtx) error {
	indieAuthError := new(domain.Error)
	if err := form.Unmarshal(ctx.QueryArgs(), r); err != nil {
		if errors.As(err, indieAuthError) {
			return indieAuthError
		}

		return domain.NewError(
			domain.ErrorCodeInvalidRequest,
			err.Error(),
			"https://indieauth.net/source/#authorization-request",
		)
	}

	if err := r.Scope.UnmarshalForm(ctx.QueryArgs().Peek("scope")); err != nil {
		if errors.As(err, indieAuthError) {
			return indieAuthError
		}

		return domain.NewError(
			domain.ErrorCodeInvalidScope,
			err.Error(),
			"https://indieweb.org/scope",
		)
	}

	if ctx.QueryArgs().Has("code_challenge") {
		err := r.CodeChallengeMethod.UnmarshalForm(ctx.QueryArgs().Peek("code_challenge_method"))
		if err != nil {
			if errors.As(err, indieAuthError) {
				return indieAuthError
			}

			return domain.NewError(
				domain.ErrorCodeInvalidRequest,
				err.Error(),
				"",
			)
		}
	}

	if err := r.ResponseType.UnmarshalForm(ctx.QueryArgs().Peek("response_type")); err != nil {
		if errors.As(err, indieAuthError) {
			return indieAuthError
		}

		return domain.NewError(
			domain.ErrorCodeUnsupportedResponseType,
			err.Error(),
			"",
		)
	}

	if r.ResponseType == domain.ResponseTypeID {
		r.ResponseType = domain.ResponseTypeCode
	}

	return nil
}

func NewAuthVerifyRequest() *AuthVerifyRequest {
	return &AuthVerifyRequest{
		Authorize:           "",
		ClientID:            new(domain.ClientID),
		CodeChallenge:       "",
		CodeChallengeMethod: domain.CodeChallengeMethodUndefined,
		Me:                  new(domain.Me),
		Provider:            "",
		RedirectURI:         new(domain.URL),
		ResponseType:        domain.ResponseTypeUndefined,
		Scope:               make(domain.Scopes, 0),
		State:               "",
	}
}

func (r *AuthVerifyRequest) bind(ctx *http.RequestCtx) error {
	indieAuthError := new(domain.Error)

	if err := form.Unmarshal(ctx.PostArgs(), r); err != nil {
		if errors.As(err, indieAuthError) {
			return indieAuthError
		}

		return domain.NewError(
			domain.ErrorCodeInvalidRequest,
			err.Error(),
			"https://indieauth.net/source/#authorization-request",
		)
	}

	if err := r.Scope.UnmarshalForm(bytes.Join(ctx.PostArgs().PeekMulti("scope[]"), []byte(" "))); err != nil {
		if errors.As(err, indieAuthError) {
			return indieAuthError
		}

		return domain.NewError(
			domain.ErrorCodeInvalidScope,
			err.Error(),
			"https://indieweb.org/scope",
		)
	}

	if ctx.PostArgs().Has("code_challenge") {
		err := r.CodeChallengeMethod.UnmarshalForm(ctx.PostArgs().Peek("code_challenge_method"))
		if err != nil {
			if errors.As(err, indieAuthError) {
				return indieAuthError
			}

			return domain.NewError(
				domain.ErrorCodeInvalidRequest,
				err.Error(),
				"",
			)
		}
	}

	if err := r.ResponseType.UnmarshalForm(ctx.PostArgs().Peek("response_type")); err != nil {
		if errors.As(err, indieAuthError) {
			return indieAuthError
		}

		return domain.NewError(
			domain.ErrorCodeUnsupportedResponseType,
			err.Error(),
			"",
		)
	}

	// NOTE(toby3d): backwards-compatible support.
	// See: https://aaronparecki.com/2020/12/03/1/indieauth-2020#response-type
	if r.ResponseType == domain.ResponseTypeID {
		r.ResponseType = domain.ResponseTypeCode
	}

	r.Provider = strings.ToLower(r.Provider)

	if !strings.EqualFold(r.Authorize, "allow") && !strings.EqualFold(r.Authorize, "deny") {
		return domain.NewError(
			domain.ErrorCodeInvalidRequest,
			"cannot validate verification request",
			"https://indieauth.net/source/#authorization-request",
		)
	}

	return nil
}

func (r *AuthExchangeRequest) bind(ctx *http.RequestCtx) error {
	indieAuthError := new(domain.Error)
	if err := form.Unmarshal(ctx.PostArgs(), r); err != nil {
		if errors.As(err, indieAuthError) {
			return indieAuthError
		}

		return domain.NewError(
			domain.ErrorCodeInvalidRequest,
			"cannot validate verification request",
			"https://indieauth.net/source/#redeeming-the-authorization-code",
		)
	}

	return nil
}
