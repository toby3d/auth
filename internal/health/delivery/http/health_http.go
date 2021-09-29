package http

import (
	"github.com/fasthttp/router"
	http "github.com/valyala/fasthttp"
)

type RequestHandler struct{}

func NewRequestHandler() *RequestHandler {
	return new(RequestHandler)
}

func (h *RequestHandler) Register(r *router.Router) {
	r.GET("/health", h.Read)
}

func (h *RequestHandler) Read(ctx *http.RequestCtx) {
	ctx.SetStatusCode(http.StatusOK)
}
