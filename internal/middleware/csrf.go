package middleware

import (
	"context"
	"crypto/subtle"
	"errors"
	"net/http"
	"time"

	"source.toby3d.me/toby3d/auth/internal/common"
	"source.toby3d.me/toby3d/auth/internal/random"
)

type (
	// CSRFConfig defines the config for CSRF middleware.
	CSRFConfig struct {
		// Skipper defines a function to skip middleware.
		Skipper Skipper

		// ErrorHandler defines a function which is executed for returning custom errors.
		ErrorHandler CSRFErrorHandler

		// TokenLookup is a string in the form of "<source>:<name>" or "<source>:<name>,<source>:<name>" that
		// is used to extract token from the request.
		// Optional. Default value "header:X-CSRF-Token".
		// Possible values:
		// - "header:<name>" or "header:<name>:<cut-prefix>"
		// - "query:<name>"
		// - "form:<name>"
		// Multiple sources example:
		// - "header:X-CSRF-Token,query:csrf"
		TokenLookup string

		// Context key to store generated CSRF token into context.
		// Optional. Default value "csrf".
		ContextKey string

		// Name of the CSRF cookie. This cookie will store CSRF token.
		// Optional. Default value "csrf".
		CookieName string

		// Domain of the CSRF cookie.
		// Optional. Default value none.
		CookieDomain string

		// Path of the CSRF cookie.
		// Optional. Default value none.
		CookiePath string

		// Max age (in seconds) of the CSRF cookie.
		// Optional. Default value 86400 (24hr).
		CookieMaxAge int

		// Indicates SameSite mode of the CSRF cookie.
		// Optional. Default value SameSiteDefaultMode.
		CookieSameSite http.SameSite

		// TokenLength is the length of the generated token.
		// Optional. Default value 32.
		TokenLength uint8

		// Indicates if CSRF cookie is secure.
		// Optional. Default value false.
		CookieSecure bool

		// Indicates if CSRF cookie is HTTP only.
		// Optional. Default value false.
		CookieHTTPOnly bool
	}

	// CSRFErrorHandler is a function which is executed for creating custom errors.
	CSRFErrorHandler func(err error, w http.ResponseWriter, r *http.Request) error
)

// ErrCSRFInvalid is returned when CSRF check fails.
var ErrCSRFInvalid = errors.New("") // echo.NewHTTPError(http.StatusForbidden, "invalid csrf token")

// DefaultCSRFConfig is the default CSRF middleware config.
var DefaultCSRFConfig = CSRFConfig{
	Skipper:        DefaultSkipper,
	TokenLength:    32,
	TokenLookup:    "header:" + common.HeaderXCSRFToken,
	ContextKey:     "csrf",
	CookieName:     "_csrf",
	CookieMaxAge:   86400,
	CookieSameSite: http.SameSiteDefaultMode,
}

// CSRF returns a Cross-Site Request Forgery (CSRF) middleware.
// See: https://en.wikipedia.org/wiki/Cross-site_request_forgery
func CSRF() Interceptor {
	c := DefaultCSRFConfig

	return CSRFWithConfig(c)
}

// CSRFWithConfig returns a CSRF middleware with config.
// See `CSRF()`.
func CSRFWithConfig(config CSRFConfig) Interceptor {
	// Defaults
	if config.Skipper == nil {
		config.Skipper = DefaultCSRFConfig.Skipper
	}

	if config.TokenLength == 0 {
		config.TokenLength = DefaultCSRFConfig.TokenLength
	}

	if config.TokenLookup == "" {
		config.TokenLookup = DefaultCSRFConfig.TokenLookup
	}

	if config.ContextKey == "" {
		config.ContextKey = DefaultCSRFConfig.ContextKey
	}

	if config.CookieName == "" {
		config.CookieName = DefaultCSRFConfig.CookieName
	}

	if config.CookieMaxAge == 0 {
		config.CookieMaxAge = DefaultCSRFConfig.CookieMaxAge
	}

	if config.CookieSameSite == http.SameSiteNoneMode {
		config.CookieSecure = true
	}

	extractors, err := CreateExtractors(config.TokenLookup)
	if err != nil {
		panic(err)
	}

	return func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		if config.Skipper(w, r) {
			next(w, r)

			return
		}

		token := ""
		if k, err := r.Cookie(config.CookieName); err != nil {
			token, _ = random.String(config.TokenLength) // Generate token
		} else {
			token = k.Value // Reuse token
		}

		switch r.Method {
		case "", http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace:
		default:
			var lastExtractorErr, lastTokenErr error
		outer:
			for _, extractor := range extractors {
				clientTokens, err := extractor(w, r)
				if err != nil {
					lastExtractorErr = err

					continue
				}

				for _, clientToken := range clientTokens {
					if validateCSRFToken(token, clientToken) {
						lastTokenErr = nil
						lastExtractorErr = nil

						break outer
					}

					lastTokenErr = ErrCSRFInvalid
				}
			}

			var finalErr error

			if lastTokenErr != nil {
				finalErr = lastTokenErr
			} else if lastExtractorErr != nil {
				switch {
				case errors.Is(lastExtractorErr, errQueryExtractorValueMissing):
					lastExtractorErr = errors.New("missing csrf token in the query string")
				case errors.Is(lastExtractorErr, errFormExtractorValueMissing):
					lastExtractorErr = errors.New("missing csrf token in the form parameter")
				case errors.Is(lastExtractorErr, errHeaderExtractorValueMissing):
					lastExtractorErr = errors.New("missing csrf token in request header")
				}

				finalErr = lastExtractorErr
			}

			if finalErr != nil {
				if config.ErrorHandler != nil {
					config.ErrorHandler(finalErr, w, r)

					return
				}

				http.Error(w, finalErr.Error(), http.StatusBadRequest)

				return
			}
		}

		// Set CSRF cookie
		cookie := new(http.Cookie)
		cookie.Name = config.CookieName
		cookie.Value = token

		if config.CookiePath != "" {
			cookie.Path = config.CookiePath
		}

		if config.CookieDomain != "" {
			cookie.Domain = config.CookieDomain
		}

		if config.CookieSameSite != http.SameSiteDefaultMode {
			cookie.SameSite = config.CookieSameSite
		}

		cookie.Expires = time.Now().Add(time.Duration(config.CookieMaxAge) * time.Second)
		cookie.Secure = config.CookieSecure
		cookie.HttpOnly = config.CookieHTTPOnly

		http.SetCookie(w, cookie)

		// Store token in the context
		r = r.WithContext(context.WithValue(r.Context(), config.ContextKey, token))

		// Protect clients from caching the response
		w.Header().Add(common.HeaderVary, common.HeaderCookie)

		next(w, r)
	}
}

func validateCSRFToken(token, clientToken string) bool {
	return subtle.ConstantTimeCompare([]byte(token), []byte(clientToken)) == 1
}
