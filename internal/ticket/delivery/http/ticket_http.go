package http

import (
	"errors"
	"fmt"
	"path"

	"github.com/fasthttp/router"
	"github.com/goccy/go-json"
	"github.com/lestrrat-go/jwx/jwa"
	http "github.com/valyala/fasthttp"
	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"source.toby3d.me/toby3d/form"
	"source.toby3d.me/toby3d/middleware"
	"source.toby3d.me/website/indieauth/internal/common"
	"source.toby3d.me/website/indieauth/internal/domain"
	"source.toby3d.me/website/indieauth/internal/random"
	"source.toby3d.me/website/indieauth/internal/ticket"
	"source.toby3d.me/website/indieauth/web"
)

type (
	TicketGenerateRequest struct {
		// The access token should be used when acting on behalf of this URL.
		Subject *domain.Me `form:"subject"`

		// The access token will work at this URL.
		Resource *domain.URL `form:"resource"`
	}

	TicketExchangeRequest struct {
		// A random string that can be redeemed for an access token.
		Ticket string `form:"ticket"`

		// The access token should be used when acting on behalf of this URL.
		Subject *domain.Me `form:"subject"`

		// The access token will work at this URL.
		Resource *domain.URL `form:"resource"`
	}

	RequestHandler struct {
		config  *domain.Config
		matcher language.Matcher
		tickets ticket.UseCase
	}
)

func NewRequestHandler(tickets ticket.UseCase, matcher language.Matcher, config *domain.Config) *RequestHandler {
	return &RequestHandler{
		config:  config,
		matcher: matcher,
		tickets: tickets,
	}
}

func (h *RequestHandler) Register(r *router.Router) {
	chain := middleware.Chain{
		middleware.CSRFWithConfig(middleware.CSRFConfig{
			Skipper: func(ctx *http.RequestCtx) bool {
				matched, _ := path.Match("/ticket*", string(ctx.Path()))

				return ctx.IsPost() && matched
			},
			CookieMaxAge:   0,
			CookieSameSite: http.CookieSameSiteStrictMode,
			ContextKey:     "csrf",
			CookieDomain:   h.config.Server.Domain,
			CookieName:     "__Secure-csrf",
			CookiePath:     "",
			TokenLookup:    "form:_csrf",
			TokenLength:    0,
			CookieSecure:   true,
			CookieHTTPOnly: true,
		}),
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

	r.GET("/ticket", chain.RequestHandler(h.handleRender))
	r.POST("/api/ticket", chain.RequestHandler(h.handleSend))
	r.POST("/ticket", chain.RequestHandler(h.handleRedeem))
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

	csrf, _ := ctx.UserValue("csrf").([]byte)
	web.WriteTemplate(ctx, &web.TicketPage{
		BaseOf: baseOf,
		CSRF:   csrf,
	})
}

func (h *RequestHandler) handleSend(ctx *http.RequestCtx) {
	ctx.Response.Header.Set(http.HeaderAccessControlAllowOrigin, h.config.Server.Domain)
	ctx.SetContentType(common.MIMEApplicationJSONCharsetUTF8)
	ctx.SetStatusCode(http.StatusOK)

	encoder := json.NewEncoder(ctx)

	req := new(TicketGenerateRequest)
	if err := req.bind(ctx); err != nil {
		ctx.SetStatusCode(http.StatusBadRequest)

		_ = encoder.Encode(err)

		return
	}

	ticket := &domain.Ticket{
		Ticket:   "",
		Resource: req.Resource,
		Subject:  req.Subject,
	}

	var err error
	if ticket.Ticket, err = random.String(h.config.TicketAuth.Length); err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)

		_ = encoder.Encode(domain.NewError(domain.ErrorCodeServerError, err.Error(), ""))

		return
	}

	if err = h.tickets.Generate(ctx, ticket); err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)

		_ = encoder.Encode(domain.NewError(domain.ErrorCodeServerError, err.Error(), ""))

		return
	}

	ctx.SetStatusCode(http.StatusOK)
}

func (h *RequestHandler) handleRedeem(ctx *http.RequestCtx) {
	ctx.SetContentType(common.MIMEApplicationJSONCharsetUTF8)
	ctx.SetStatusCode(http.StatusOK)

	encoder := json.NewEncoder(ctx)

	req := new(TicketExchangeRequest)
	if err := req.bind(ctx); err != nil {
		ctx.SetStatusCode(http.StatusBadRequest)

		_ = encoder.Encode(err)

		return
	}

	token, err := h.tickets.Redeem(ctx, &domain.Ticket{
		Ticket:   req.Ticket,
		Resource: req.Resource,
		Subject:  req.Subject,
	})
	if err != nil {
		ctx.SetStatusCode(http.StatusBadRequest)

		_ = encoder.Encode(domain.NewError(domain.ErrorCodeServerError, err.Error(), ""))

		return
	}

	// TODO(toby3d): print the result as part of the debugging. Instead, we
	// need to send or save the token to the recipient for later use.
	ctx.SetBodyString(fmt.Sprintf(`{
		"access_token": "%s",
		"token_type": "Bearer",
		"scope": "%s",
		"me": "%s"
	}`, token.AccessToken, token.Scope.String(), token.Me.String()))
}

func (req *TicketGenerateRequest) bind(ctx *http.RequestCtx) (err error) {
	indieAuthError := new(domain.Error)
	if err = form.Unmarshal(ctx.Request.PostArgs(), req); err != nil {
		if errors.As(err, indieAuthError) {
			return indieAuthError
		}

		return domain.NewError(
			domain.ErrorCodeInvalidRequest,
			err.Error(),
			"https://indieweb.org/IndieAuth_Ticket_Auth#Create_the_IndieAuth_ticket",
		)
	}

	if req.Resource == nil {
		return domain.NewError(
			domain.ErrorCodeInvalidRequest,
			"resource value MUST be set",
			"https://indieweb.org/IndieAuth_Ticket_Auth#Create_the_IndieAuth_ticket",
		)
	}

	if req.Subject == nil {
		return domain.NewError(
			domain.ErrorCodeInvalidRequest,
			"subject value MUST be set",
			"https://indieweb.org/IndieAuth_Ticket_Auth#Create_the_IndieAuth_ticket",
		)
	}

	return nil
}

func (req *TicketExchangeRequest) bind(ctx *http.RequestCtx) (err error) {
	indieAuthError := new(domain.Error)
	if err = form.Unmarshal(ctx.Request.PostArgs(), req); err != nil {
		if errors.As(err, indieAuthError) {
			return indieAuthError
		}

		return domain.NewError(
			domain.ErrorCodeInvalidRequest,
			err.Error(),
			"https://indieweb.org/IndieAuth_Ticket_Auth#Create_the_IndieAuth_ticket",
		)
	}

	if req.Ticket == "" {
		return domain.NewError(
			domain.ErrorCodeInvalidRequest,
			"ticket parameter is required",
			"https://indieweb.org/IndieAuth_Ticket_Auth#Create_the_IndieAuth_ticket",
		)
	}

	if req.Resource == nil {
		return domain.NewError(
			domain.ErrorCodeInvalidRequest,
			"resource parameter is required",
			"https://indieweb.org/IndieAuth_Ticket_Auth#Create_the_IndieAuth_ticket",
		)
	}

	if req.Subject == nil {
		return domain.NewError(
			domain.ErrorCodeInvalidRequest,
			"subject parameter is required",
			"https://indieweb.org/IndieAuth_Ticket_Auth#Create_the_IndieAuth_ticket",
		)
	}

	return nil
}
