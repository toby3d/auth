package http

import (
	"github.com/fasthttp/router"
	http "github.com/valyala/fasthttp"

	"source.toby3d.me/website/oauth/internal/common"
)

type RequestHandler struct{}

func NewRequestHandler() *RequestHandler {
	return new(RequestHandler)
}

func (h *RequestHandler) Register(r *router.Router) {
	r.GET("/health", h.RequestHandler)
}

func (h *RequestHandler) RequestHandler(ctx *http.RequestCtx) {
	ctx.SetContentType(common.MIMETextPlainCharsetUTF8)
	ctx.SetStatusCode(http.StatusOK)
}
