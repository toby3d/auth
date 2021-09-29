package middleware

import http "github.com/valyala/fasthttp"

type (
	Chain []Interceptor

	Interceptor func(*http.RequestCtx, http.RequestHandler)

	RequestHandler http.RequestHandler

	Skipper func(*http.RequestCtx) bool
)

// DefaultSkipper is the default skipper, which always returns false.
//nolint: gochecknoglobals
var DefaultSkipper Skipper = func(*http.RequestCtx) bool { return false }

func (count RequestHandler) Intercept(middleware Interceptor) RequestHandler {
	return func(ctx *http.RequestCtx) { middleware(ctx, http.RequestHandler(count)) }
}

func (chain Chain) RequestHandler(handler http.RequestHandler) http.RequestHandler {
	current := RequestHandler(handler)

	for i := len(chain) - 1; i >= 0; i-- {
		m := chain[i]
		current = current.Intercept(m)
	}

	return http.RequestHandler(current)
}
