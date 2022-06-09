package middleware

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"time"

	http "github.com/valyala/fasthttp"
)

// CSRFConfig defines the config for CSRF middleware.
type CSRFConfig struct {
	// Skipper defines a function to skip middleware.
	Skipper Skipper

	// Indicates SameSite mode of the CSRF cookie.
	//
	// Optional. Default value SameSiteDefaultMode.
	CookieSameSite http.CookieSameSite

	// Context key to store generated CSRF token into context.
	//
	// Optional. Default value "csrf".
	ContextKey string

	// Domain of the CSRF cookie.
	//
	// Optional. Default value none.
	CookieDomain string

	// Name of the CSRF cookie. This cookie will store CSRF token.
	//
	// Optional. Default value "csrf".
	CookieName string

	// Path of the CSRF cookie.
	//
	// Optional. Default value none.
	CookiePath string

	// TokenLookup is a string in the form of "<source>:<key>" that is used
	// to extract token from the request.
	// Possible values:
	// - "header:<name>"
	// - "form:<name>"
	// - "query:<name>"
	//
	// Optional. Default value "header:X-CSRF-Token".
	TokenLookup string

	// Max age (in seconds) of the CSRF cookie.
	//
	// Optional. Default value 86400 (24hr).
	CookieMaxAge int

	// TokenLength is the length of the generated token.
	//
	// Optional. Default value 32.
	TokenLength int

	// Indicates if CSRF cookie is secure.
	//
	// Optional. Default value false.
	CookieSecure bool

	// Indicates if CSRF cookie is HTTP only.
	//
	// Optional. Default value false.
	CookieHTTPOnly bool
}

const (
	ExtractorLimit int = 20

	// HeaderXCSRFToken describes the name of the header with the CSRF token.
	HeaderXCSRFToken string = "X-CSRF-Token"
)

var (
	ErrCSRFInvalid = errors.New("invalid csrf token")

	// DefaultCSRFConfig contains the default CSRF middleware configuration.
	//nolint: gochecknoglobals, gomnd
	DefaultCSRFConfig = CSRFConfig{
		Skipper:        DefaultSkipper,
		CookieMaxAge:   86400,
		CookieSameSite: http.CookieSameSiteDefaultMode,
		ContextKey:     "csrf",
		CookieDomain:   "",
		CookieName:     "_csrf",
		CookiePath:     "",
		TokenLookup:    "header:" + HeaderXCSRFToken,
		TokenLength:    32,
		CookieSecure:   true,
		CookieHTTPOnly: false,
	}
)

// CSRF returns a Cross-Site Request Forgery (CSRF) middleware.
func CSRF() Interceptor {
	config := DefaultCSRFConfig

	return CSRFWithConfig(config)
}

// CSRFWithConfig returns a CSRF middleware with config.
//nolint: funlen, cyclop
func CSRFWithConfig(config CSRFConfig) Interceptor {
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

	if config.CookieSameSite == http.CookieSameSiteNoneMode {
		config.CookieSecure = DefaultCSRFConfig.CookieSecure
	}

	extractors, err := createExtractors(config.TokenLookup, "")
	if err != nil {
		panic("middleware: csrf: " + err.Error())
	}

	return func(ctx *http.RequestCtx, next http.RequestHandler) {
		if config.Skipper(ctx) {
			next(ctx)

			return
		}

		var token []byte
		if k := ctx.Request.Header.Cookie(config.CookieName); k != nil {
			token = k
		} else {
			token = make([]byte, config.TokenLength)
			if _, err := rand.Read(token); err != nil {
				ctx.Error(err.Error(), http.StatusInternalServerError)

				return
			}

			token = []byte(base64.RawURLEncoding.EncodeToString(token))
		}

		switch {
		case ctx.IsGet(), ctx.IsHead(), ctx.IsOptions(), ctx.IsTrace():
		default:
			var lastExtractorErr, lastTokenErr error

		outer:
			for _, extractor := range extractors {
				clientTokens, err := extractor(ctx)
				if err != nil {
					lastExtractorErr = err

					continue
				}

				for _, clientToken := range clientTokens {
					if !validateCSRFToken(token, clientToken) {
						lastTokenErr = ErrCSRFInvalid

						continue
					}

					lastTokenErr, lastExtractorErr = nil, nil

					break outer
				}
			}

			if lastTokenErr != nil {
				ctx.Error(lastTokenErr.Error(), http.StatusInternalServerError)

				return
			} else if lastExtractorErr != nil {
				ctx.Error(lastExtractorErr.Error(), http.StatusBadRequest)

				return
			}
		}

		// NOTE(toby3d): set CSRF cookie
		cookie := http.AcquireCookie()
		defer http.ReleaseCookie(cookie)

		cookie.SetKey(config.CookieName)
		cookie.SetValueBytes(token)

		if config.CookiePath != "" {
			cookie.SetPath(config.CookiePath)
		}

		if config.CookieDomain != "" {
			cookie.SetDomain(config.CookieDomain)
		}

		if config.CookieSameSite != http.CookieSameSiteDefaultMode {
			cookie.SetSameSite(config.CookieSameSite)
		}

		cookie.SetExpire(time.Now().Add(time.Duration(config.CookieMaxAge) * time.Second))
		cookie.SetSecure(config.CookieSecure)
		cookie.SetHTTPOnly(config.CookieHTTPOnly)
		ctx.Response.Header.SetCookie(cookie)

		// NOTE(toby3d): store token in the context
		ctx.SetUserValue(config.ContextKey, token)

		// NOTE(toby3d): protect clients from caching the response
		ctx.Response.Header.Add(http.HeaderVary, http.HeaderCookie)

		next(ctx)
	}
}

func validateCSRFToken(token, clientToken []byte) bool {
	return subtle.ConstantTimeCompare(token, clientToken) == 1
}
