// Package middleware provides methods to analyze and mutate data delivered over
// HTTP.
package middleware

import http "github.com/valyala/fasthttp"

type (
	// BeforeFunc is an alias of fasthttp.RequestHandler to execute its own
	// logic before middleware.
	BeforeFunc = http.RequestHandler

	// Chain describes the middlewares call chain.
	Chain []Interceptor

	// Interceptor describes the middleware contract.
	Interceptor func(*http.RequestCtx, http.RequestHandler)

	// RequestHandler is an alias of fasthttp.RequestHandler for the
	// extension.
	RequestHandler http.RequestHandler

	// Skipper defines a function to skip middleware. Returning true skips
	// processing the middleware.
	Skipper func(*http.RequestCtx) bool
)

// DefaultSkipper is the default skipper, which always returns false.
//nolint: gochecknoglobals
var DefaultSkipper Skipper = func(*http.RequestCtx) bool { return false }

// Intercept wraps the RequestHandler in middleware input.
func (count RequestHandler) Intercept(middleware Interceptor) RequestHandler {
	return func(ctx *http.RequestCtx) {
		middleware(ctx, http.RequestHandler(count))
	}
}

// RequestHandler wraps the input into the middlewares chain starting from the
// last to first and returns fasthttp.RequestHandler.
func (chain Chain) RequestHandler(handler http.RequestHandler) http.RequestHandler {
	current := RequestHandler(handler)

	for i := len(chain) - 1; i >= 0; i-- {
		m := chain[i]
		current = current.Intercept(m)
	}

	return http.RequestHandler(current)
}
