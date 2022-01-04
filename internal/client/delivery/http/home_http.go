package http

import (
	"encoding/base64"
	"strings"

	"github.com/fasthttp/router"
	http "github.com/valyala/fasthttp"
	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"source.toby3d.me/website/indieauth/internal/common"
	"source.toby3d.me/website/indieauth/internal/domain"
	"source.toby3d.me/website/indieauth/internal/random"
	"source.toby3d.me/website/indieauth/web"
)

type RequestHandler struct {
	client  *domain.Client
	config  *domain.Config
	matcher language.Matcher
}

const DefaultStateLength int = 64

func NewRequestHandler(config *domain.Config, client *domain.Client, matcher language.Matcher) *RequestHandler {
	return &RequestHandler{
		client:  client,
		config:  config,
		matcher: matcher,
	}
}

func (h *RequestHandler) Register(r *router.Router) {
	r.GET("/", h.read)
}

func (h *RequestHandler) read(ctx *http.RequestCtx) {
	ctx.SetContentType(common.MIMETextHTMLCharsetUTF8)

	// TODO(toby3d): save state for checking it at the end of flow?
	state, err := random.Bytes(DefaultStateLength)
	if err != nil {
		ctx.Error(err.Error(), http.StatusInternalServerError)

		return
	}

	redirectUri := make([]string, len(h.client.RedirectURI))
	for i := range h.client.RedirectURI {
		redirectUri[i] = h.client.RedirectURI[i].String()
	}

	tags, _, _ := language.ParseAcceptLanguage(string(ctx.Request.Header.Peek(http.HeaderAcceptLanguage)))
	tag, _, _ := h.matcher.Match(tags...)

	ctx.Response.Header.Set(
		http.HeaderLink, `<`+strings.Join(redirectUri, `>; rel="redirect_uri", `)+`>; rel="redirect_uri"`,
	)
	web.WriteTemplate(ctx, &web.HomePage{
		BaseOf: web.BaseOf{
			Config:   h.config,
			Language: tag,
			Printer:  message.NewPrinter(tag),
		},
		RedirectURI:  h.client.RedirectURI,
		AuthEndpoint: "/authorize",
		Client:       h.client,
		State:        base64.RawURLEncoding.EncodeToString(state),
	})
}
