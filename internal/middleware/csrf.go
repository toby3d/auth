package middleware

import (
	"crypto/subtle"
	"errors"
	"strings"
	"time"

	http "github.com/valyala/fasthttp"

	"source.toby3d.me/website/oauth/internal/random"
)

type (
	CSRFConfig struct {
		Skipper        Skipper
		CookieMaxAge   time.Duration
		CookieSameSite http.CookieSameSite
		ContextKey     string
		CookieDomain   string
		CookieName     string
		CookiePath     string
		TokenLookup    string
		TokenLength    int
		CookieSecure   bool
		CookieHTTPOnly bool
	}

	csrfTokenExtractor func(*http.RequestCtx) ([]byte, error)
)

// HeaderXCSRFToken describes the name of the header with the CSRF token.
//nolint: gosec
const HeaderXCSRFToken string = "X-CSRF-Token"

var (
	ErrMissingFormToken  = errors.New("missing csrf token in the form parameter")
	ErrMissingQueryToken = errors.New("missing csrf token in the query string")
)

// DefaultCSRFConfig contains the default CSRF middleware configuration.
//nolint: exhaustivestruct, gochecknoglobals, gomnd
var DefaultCSRFConfig = CSRFConfig{
	ContextKey:     "csrf",
	CookieHTTPOnly: false,
	CookieMaxAge:   24 * time.Hour,
	CookieName:     "_csrf",
	CookieSameSite: http.CookieSameSiteDefaultMode,
	CookieSecure:   false,
	Skipper:        DefaultSkipper,
	TokenLength:    32,
	TokenLookup:    "header:" + HeaderXCSRFToken,
}

func CSRF() Interceptor {
	cfg := DefaultCSRFConfig

	return CSRFWithConfig(cfg)
}

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

		var token []byte
		if k := ctx.Request.Header.Cookie(config.CookieName); k != nil {
			token = k
		} else {
			token = []byte(random.New().String(config.TokenLength))
		}

		switch string(ctx.Method()) {
		case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace:
		default:
			// NOTE(toby3d): validate token only for requests which are not defined as 'safe' by RFC7231
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

		cookie.SetExpire(time.Now().Add(config.CookieMaxAge))
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

func csrfTokenFromHeader(header string) csrfTokenExtractor {
	return func(ctx *http.RequestCtx) ([]byte, error) {
		return ctx.Request.Header.Peek(header), nil
	}
}

func csrfTokenFromForm(param string) csrfTokenExtractor {
	return func(ctx *http.RequestCtx) ([]byte, error) {
		token := ctx.FormValue(param)
		if token == nil {
			return nil, ErrMissingFormToken
		}

		return token, nil
	}
}

func csrfTokenFromQuery(param string) csrfTokenExtractor {
	return func(ctx *http.RequestCtx) ([]byte, error) {
		if !ctx.QueryArgs().Has(param) {
			return nil, ErrMissingQueryToken
		}

		return ctx.QueryArgs().Peek(param), nil
	}
}

func validateCSRFToken(token, clientToken []byte) bool {
	return subtle.ConstantTimeCompare(token, clientToken) == 1
}
