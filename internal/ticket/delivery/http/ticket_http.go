package http

import (
	"fmt"
	"path"

	"github.com/fasthttp/router"
	"github.com/goccy/go-json"
	http "github.com/valyala/fasthttp"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/xerrors"

	"source.toby3d.me/toby3d/form"
	"source.toby3d.me/toby3d/middleware"
	"source.toby3d.me/website/indieauth/internal/common"
	"source.toby3d.me/website/indieauth/internal/domain"
	"source.toby3d.me/website/indieauth/internal/random"
	"source.toby3d.me/website/indieauth/internal/ticket"
	"source.toby3d.me/website/indieauth/web"
)

type (
	GenerateRequest struct {
		// The access token should be used when acting on behalf of this URL.
		Subject *domain.URL `form:"subject"`

		// The access token will work at this URL.
		Resource *domain.URL `form:"resource"`
	}

	ExchangeRequest struct {
		// A random string that can be redeemed for an access token.
		Ticket string `form:"ticket"`

		// The access token should be used when acting on behalf of this URL.
		Subject *domain.URL `form:"subject"`

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
			CookieSameSite: http.CookieSameSiteLaxMode,
			ContextKey:     "csrf",
			CookieName:     "_csrf",
			TokenLookup:    "form:_csrf",
			CookieSecure:   true,
			CookieHTTPOnly: true,
			Skipper: func(ctx *http.RequestCtx) bool {
				matched, _ := path.Match("/ticket*", string(ctx.Path()))

				return ctx.IsPost() && matched
			},
		}),
		middleware.LogFmt(),
	}
	// TODO(toby3d): secure this via JWT middleware
	r.GET("/ticket", chain.RequestHandler(h.handleRender))
	r.POST("/api/ticket", chain.RequestHandler(h.handleSend))
	r.POST("/ticket", chain.RequestHandler(h.handleExchange))
}

func (h *RequestHandler) handleRender(ctx *http.RequestCtx) {
	ctx.SetContentType(common.MIMETextHTMLCharsetUTF8)

	tags, _, _ := language.ParseAcceptLanguage(string(ctx.Request.Header.Peek(http.HeaderAcceptLanguage)))
	tag, _, _ := h.matcher.Match(tags...)

	csrf, _ := ctx.UserValue("csrf").([]byte)

	ctx.SetContentType(common.MIMETextHTMLCharsetUTF8)
	web.WriteTemplate(ctx, &web.TicketPage{
		BaseOf: web.BaseOf{
			Config:   h.config,
			Language: tag,
			Printer:  message.NewPrinter(tag),
		},
		CSRF: csrf,
	})
}

func (h *RequestHandler) handleSend(ctx *http.RequestCtx) {
	ctx.SetContentType(common.MIMEApplicationJSONCharsetUTF8)
	ctx.SetStatusCode(http.StatusOK)

	encoder := json.NewEncoder(ctx)

	req := new(ExchangeRequest)
	if err := req.bind(ctx); err != nil {
		ctx.SetStatusCode(http.StatusBadRequest)
		encoder.Encode(err)

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
		encoder.Encode(&domain.Error{
			Code:        "unauthorized_client",
			Description: err.Error(),
			Frame:       xerrors.Caller(1),
		})

		return
	}

	if err = h.tickets.Generate(ctx, ticket); err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		encoder.Encode(&domain.Error{
			Code:        "unauthorized_client",
			Description: err.Error(),
			Frame:       xerrors.Caller(1),
		})

		return
	}

	ctx.SetStatusCode(http.StatusOK)
}

func (h *RequestHandler) handleExchange(ctx *http.RequestCtx) {
	ctx.SetContentType(common.MIMEApplicationJSONCharsetUTF8)
	ctx.SetStatusCode(http.StatusOK)

	encoder := json.NewEncoder(ctx)

	req := new(ExchangeRequest)
	if err := req.bind(ctx); err != nil {
		ctx.SetStatusCode(http.StatusBadRequest)
		encoder.Encode(err)

		return
	}

	token, err := h.tickets.Exchange(ctx, &domain.Ticket{
		Ticket:   req.Ticket,
		Resource: req.Resource,
		Subject:  req.Subject,
	})
	if err != nil {
		ctx.SetStatusCode(http.StatusBadRequest)
		encoder.Encode(domain.Error{
			Code:        "invalid_request",
			Description: err.Error(),
			Frame:       xerrors.Caller(1),
		})

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

func (req *GenerateRequest) bind(ctx *http.RequestCtx) (err error) {
	if err = form.Unmarshal(ctx.Request.PostArgs(), req); err != nil {
		return domain.Error{
			Code:        "invalid_request",
			Description: err.Error(),
			URI:         "https://indieweb.org/IndieAuth_Ticket_Auth#Create_the_IndieAuth_ticket",
			Frame:       xerrors.Caller(1),
		}
	}

	if req.Resource == nil {
		return domain.Error{
			Code:        "invalid_request",
			Description: "resource value MUST be set",
			URI:         "https://indieweb.org/IndieAuth_Ticket_Auth#Create_the_IndieAuth_ticket",
			Frame:       xerrors.Caller(1),
		}
	}

	if req.Subject == nil {
		return domain.Error{
			Code:        "invalid_request",
			Description: "subject value MUST be set",
			URI:         "https://indieweb.org/IndieAuth_Ticket_Auth#Create_the_IndieAuth_ticket",
			Frame:       xerrors.Caller(1),
		}
	}

	return nil
}

func (req *ExchangeRequest) bind(ctx *http.RequestCtx) (err error) {
	if err = form.Unmarshal(ctx.Request.PostArgs(), req); err != nil {
		return domain.Error{
			Code:        "invalid_request",
			Description: err.Error(),
			URI:         "https://indieweb.org/IndieAuth_Ticket_Auth#Create_the_IndieAuth_ticket",
			Frame:       xerrors.Caller(1),
		}
	}

	if req.Ticket == "" {
		return domain.Error{
			Code:        "invalid_request",
			Description: "ticket parameter is required",
			URI:         "https://indieweb.org/IndieAuth_Ticket_Auth#Create_the_IndieAuth_ticket",
			Frame:       xerrors.Caller(1),
		}
	}

	if req.Resource == nil {
		return domain.Error{
			Code:        "invalid_request",
			Description: "resource value MUST be set",
			URI:         "https://indieweb.org/IndieAuth_Ticket_Auth#Create_the_IndieAuth_ticket",
			Frame:       xerrors.Caller(1),
		}
	}

	if req.Subject == nil {
		return domain.Error{
			Code:        "invalid_request",
			Description: "subject value MUST be set",
			URI:         "https://indieweb.org/IndieAuth_Ticket_Auth#Create_the_IndieAuth_ticket",
			Frame:       xerrors.Caller(1),
		}
	}

	return nil
}
