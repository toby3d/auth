// Package middleware provides methods to analyze and mutate data delivered over
// HTTP.
package middleware

import (
	"net/http"
)

type (
	// BeforeFunc is an alias of http.HandlerFunc to execute its own
	// logic before middleware.
	BeforeFunc = http.HandlerFunc

	// Chain describes the middlewares call chain.
	Chain []Interceptor

	// Interceptor describes the middleware contract.
	Interceptor func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc)

	// HandlerFunc is an alias of http.HandlerFunc for the
	// extension.
	HandlerFunc http.HandlerFunc

	// Skipper defines a function to skip middleware. Returning true skips
	// processing the middleware.
	Skipper func(w http.ResponseWriter, r *http.Request) bool
)

// DefaultSkipper is the default skipper, which always returns false.
//
//nolint:gochecknoglobals
var DefaultSkipper Skipper = func(w http.ResponseWriter, r *http.Request) bool { return false }

// Intercept wraps the HandlerFunc in middleware input.
func (count HandlerFunc) Intercept(middleware Interceptor) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		middleware(w, r, http.HandlerFunc(count))
	}
}

// Handler wraps the input into the middlewares chain starting from the
// last to first and returns http.Handler.
func (chain Chain) Handler(handler http.HandlerFunc) http.Handler {
	current := HandlerFunc(handler)

	for i := len(chain) - 1; i >= 0; i-- {
		m := chain[i]
		current = current.Intercept(m)
	}

	return http.HandlerFunc(current)
}
