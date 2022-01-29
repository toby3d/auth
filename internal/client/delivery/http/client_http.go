package http

import (
	"errors"
	"strings"

	"github.com/fasthttp/router"
	http "github.com/valyala/fasthttp"
	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"source.toby3d.me/toby3d/form"
	"source.toby3d.me/toby3d/middleware"
	"source.toby3d.me/website/indieauth/internal/common"
	"source.toby3d.me/website/indieauth/internal/domain"
	"source.toby3d.me/website/indieauth/internal/token"
	"source.toby3d.me/website/indieauth/web"
)

type (
	CallbackRequest struct {
		Iss              *domain.ClientID `form:"iss"`
		Code             string           `form:"code"`
		Error            string           `form:"error"`
		ErrorDescription string           `form:"error_description"`
		State            string           `form:"state"`
	}

	NewRequestHandlerOptions struct {
		Client  *domain.Client
		Config  *domain.Config
		Matcher language.Matcher
		Tokens  token.UseCase
	}

	RequestHandler struct {
		client  *domain.Client
		config  *domain.Config
		matcher language.Matcher
		tokens  token.UseCase
	}
)

func NewRequestHandler(opts NewRequestHandlerOptions) *RequestHandler {
	return &RequestHandler{
		client:  opts.Client,
		config:  opts.Config,
		matcher: opts.Matcher,
		tokens:  opts.Tokens,
	}
}

func (h *RequestHandler) Register(r *router.Router) {
	chain := middleware.Chain{
		middleware.LogFmt(),
	}

	r.GET("/", chain.RequestHandler(h.handleRender))
	r.GET("/callback", chain.RequestHandler(h.handleCallback))
}

func (h *RequestHandler) handleRender(ctx *http.RequestCtx) {
	redirectUri := make([]string, len(h.client.RedirectURI))
	for i := range h.client.RedirectURI {
		redirectUri[i] = h.client.RedirectURI[i].String()
	}

	ctx.Response.Header.Set(
		http.HeaderLink, `<`+strings.Join(redirectUri, `>; rel="redirect_uri", `)+`>; rel="redirect_uri"`,
	)

	tags, _, _ := language.ParseAcceptLanguage(string(ctx.Request.Header.Peek(http.HeaderAcceptLanguage)))
	tag, _, _ := h.matcher.Match(tags...)

	// TODO(toby3d): generate and store PKCE

	ctx.SetContentType(common.MIMETextHTMLCharsetUTF8)
	web.WriteTemplate(ctx, &web.HomePage{
		BaseOf: web.BaseOf{
			Config:   h.config,
			Language: tag,
			Printer:  message.NewPrinter(tag),
		},
		Client: h.client,
		State:  "hackme", // TODO(toby3d): generate and store state
	})
}

func (h *RequestHandler) handleCallback(ctx *http.RequestCtx) {
	req := new(CallbackRequest)
	if err := req.bind(ctx); err != nil {
		ctx.Error(err.Error(), http.StatusInternalServerError)

		return
	}

	if req.Error != "" {
		ctx.Error(req.ErrorDescription, http.StatusUnauthorized)

		return
	}

	// TODO(toby3d): load and check state

	if req.Iss.String() != h.client.ID.String() {
		ctx.Error("iss is not equal", http.StatusBadRequest)

		return
	}

	token, _, err := h.tokens.Exchange(ctx, token.ExchangeOptions{
		ClientID:     h.client.ID,
		RedirectURI:  h.client.RedirectURI[0],
		Code:         req.Code,
		CodeVerifier: "", // TODO(toby3d): validate PKCE here
	})
	if err != nil {
		ctx.Error(err.Error(), http.StatusBadRequest)

		return
	}

	tags, _, _ := language.ParseAcceptLanguage(string(ctx.Request.Header.Peek(http.HeaderAcceptLanguage)))
	tag, _, _ := h.matcher.Match(tags...)

	ctx.SetContentType(common.MIMETextHTMLCharsetUTF8)
	web.WriteTemplate(ctx, &web.CallbackPage{
		BaseOf: web.BaseOf{
			Config:   h.config,
			Language: tag,
			Printer:  message.NewPrinter(tag),
		},
		Token: token,
	})
}

func (req *CallbackRequest) bind(ctx *http.RequestCtx) error {
	indieAuthError := new(domain.Error)

	if err := form.Unmarshal(ctx.QueryArgs(), req); err != nil {
		if errors.As(err, indieAuthError) {
			return indieAuthError
		}

		return domain.NewError(domain.ErrorCodeInvalidRequest, err.Error(), "")
	}

	return nil
}
