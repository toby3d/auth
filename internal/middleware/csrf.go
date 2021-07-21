package middleware

import (
	"crypto/subtle"
	"errors"
	"strings"
	"time"

	http "github.com/valyala/fasthttp"
	"gitlab.com/toby3d/indieauth/internal/random"
)

type (
	CSRFConfig struct {
		ContextKey     string
		CookieDomain   string
		CookieHTTPOnly bool
		CookieMaxAge   int
		CookieName     string
		CookiePath     string
		CookieSameSite http.CookieSameSite
		CookieSecure   bool
		Skipper        Skipper
		TokenLength    int
		TokenLookup    string
	}

	csrfTokenExtractor func(*http.RequestCtx) ([]byte, error)
)

const HeaderXCSRFToken string = "X-CSRF-Token"

var DefaultCSRFConfig = CSRFConfig{
	Skipper:        DefaultSkipper,
	TokenLength:    32,
	TokenLookup:    "header:" + HeaderXCSRFToken,
	ContextKey:     "csrf",
	CookieName:     "_csrf",
	CookieMaxAge:   86400,
	CookieSameSite: http.CookieSameSiteDefaultMode,
}

func CSRF() Interceptor {
	cfg := DefaultCSRFConfig

	return CSRFWithConfig(cfg)
}

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
		config.CookieSecure = true
	}

	parts := strings.Split(config.TokenLookup, ":")
	extractor := csrfTokenFromHeader(parts[1])

	switch parts[0] {
	case "form":
		extractor = csrfTokenFromForm(parts[1])
	case "query":
		extractor = csrfTokenFromQuery(parts[1])
	}

	return func(ctx *http.RequestCtx, next http.RequestHandler) {
		if config.Skipper(ctx) {
			next(ctx)

			return
		}

		k := ctx.Request.Header.Cookie(config.CookieName)
		var token []byte

		// Generate token
		if k == nil {
			token = []byte(random.New().String(config.TokenLength))
		} else {
			// Reuse token
			token = k
		}

		switch string(ctx.Method()) {
		case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace:
		default:
			// Validate token only for requests which are not defined as 'safe' by RFC7231
			clientToken, err := extractor(ctx)
			if err != nil {
				ctx.Error(err.Error(), http.StatusBadRequest)

				return
			}

			if !validateCSRFToken(token, clientToken) {
				ctx.Error("invalid csrf token", http.StatusForbidden)

				return
			}
		}

		// Set CSRF cookie
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

		// Store token in the context
		ctx.SetUserValue(config.ContextKey, token)

		// Protect clients from caching the response
		ctx.Response.Header.Add(http.HeaderVary, http.HeaderCookie)

		next(ctx)
	}
}

func csrfTokenFromHeader(header string) csrfTokenExtractor {
	return func(ctx *http.RequestCtx) ([]byte, error) {
		return ctx.Request.Header.Peek(header), nil
	}
}

func csrfTokenFromForm(param string) csrfTokenExtractor {
	return func(ctx *http.RequestCtx) ([]byte, error) {
		token := ctx.FormValue(param)
		if token == nil {
			return nil, errors.New("missing csrf token in the form parameter")
		}

		return token, nil
	}
}

func csrfTokenFromQuery(param string) csrfTokenExtractor {
	return func(ctx *http.RequestCtx) ([]byte, error) {
		if !ctx.QueryArgs().Has(param) {
			return nil, errors.New("missing csrf token in the query string")
		}

		return ctx.QueryArgs().Peek(param), nil
	}
}

func validateCSRFToken(token, clientToken []byte) bool {
	return subtle.ConstantTimeCompare(token, clientToken) == 1
}
