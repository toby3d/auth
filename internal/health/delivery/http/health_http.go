package http

import (
	"github.com/fasthttp/router"
	http "github.com/valyala/fasthttp"

	"source.toby3d.me/toby3d/middleware"
	"source.toby3d.me/toby3d/auth/internal/common"
)

type RequestHandler struct{}

func NewRequestHandler() *RequestHandler {
	return &RequestHandler{}
}

func (h *RequestHandler) Register(r *router.Router) {
	chain := middleware.Chain{
		middleware.LogFmt(),
	}

	r.GET("/health", chain.RequestHandler(h.read))
}

func (h *RequestHandler) read(ctx *http.RequestCtx) {
	ctx.SuccessString(common.MIMEApplicationJSONCharsetUTF8, `{"ok": true}`)
}
