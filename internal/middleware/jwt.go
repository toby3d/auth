package middleware

import (
	"bytes"
	"strings"

	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwt"
	"github.com/pkg/errors"
	http "github.com/valyala/fasthttp"
	"golang.org/x/xerrors"

	"source.toby3d.me/website/oauth/internal/domain"
)

type (
	JWTConfig struct {
		// Skipper defines a function to skip middleware.
		Skipper Skipper

		// BeforeFunc defines a function which is executed just before
		// the middleware.
		BeforeFunc BeforeFunc

		// SuccessHandler defines a function which is executed for a
		// valid token.
		SuccessHandler JWTSuccessHandler

		// ErrorHandler defines a function which is executed for an
		// invalid token. It may be used to define a custom JWT error.
		ErrorHandler JWTErrorHandler

		// ErrorHandlerWithContext is almost identical to ErrorHandler,
		// but it's passed the current context.
		ErrorHandlerWithContext JWTErrorHandlerWithContext

		// Signing key to validate token.
		// This is one of the three options to provide a token
		// validation key. The order of precedence is a user-defined
		// KeyFunc, SigningKeys and SigningKey. Required if neither
		// user-defined KeyFunc nor SigningKeys is provided.
		SigningKey interface{}

		// Map of signing keys to validate token with kid field usage.
		// This is one of the three options to provide a token
		// validation key. The order of precedence is a user-defined
		// KeyFunc, SigningKeys and SigningKey. Required if neither
		// user-defined KeyFunc nor SigningKey is provided.
		SigningKeys map[string]interface{}

		// Signing method used to check the token's signing algorithm.
		// Optional. Default value HS256.
		SigningMethod jwa.SignatureAlgorithm

		// Context key to store user information from the token into
		// context. Optional. Default value "user".
		ContextKey string

		// Claims are extendable claims data defining token content.
		// Used by default ParseTokenFunc implementation. Not used if
		// custom ParseTokenFunc is set. Optional. Default value
		// []jwt.ClaimPair
		Claims []jwt.ClaimPair

		// TokenLookup is a string in the form of "<source>:<name>" or
		// "<source>:<name>,<source>:<name>" that is used to extract
		// token from the request.
		// Optional. Default value "header:Authorization".
		// Possible values:
		// - "header:<name>"
		// - "query:<name>"
		// - "param:<name>"
		// - "cookie:<name>"
		// - "form:<name>"
		// Multiply sources example:
		// - "header: Authorization,cookie: myowncookie"
		TokenLookup string

		// AuthScheme to be used in the Authorization header. Optional.
		// Default value "Bearer".
		AuthScheme string

		// KeyFunc defines a user-defined function that supplies the
		// public key for a token validation. The function shall take
		// care of verifying the signing algorithm and selecting the
		// proper key. A user-defined KeyFunc can be useful if tokens
		// are issued by an external party. Used by default
		// ParseTokenFunc implementation.
		//
		// When a user-defined KeyFunc is provided, SigningKey,
		// SigningKeys, and SigningMethod are ignored. This is one of
		// the three options to provide a token validation key. The
		// order of precedence is a user-defined KeyFunc, SigningKeys
		// and SigningKey. Required if neither SigningKeys nor
		// SigningKey is provided. Not used if custom ParseTokenFunc is
		// set. Default to an internal implementation verifying the
		// signing algorithm and selecting the proper key.
		// KeyFunc jwt.Keyfunc

		// ParseTokenFunc defines a user-defined function that parses
		// token from given auth. Returns an error when token parsing
		// fails or parsed token is invalid. Defaults to implementation
		// using `github.com/golang-jwt/jwt` as JWT implementation library
		ParseTokenFunc func(auth []byte, ctx *http.RequestCtx) (interface{}, error)
	}

	// JWTSuccessHandler defines a function which is executed for a valid
	// token.
	JWTSuccessHandler func(ctx *http.RequestCtx)

	// JWTErrorHandler defines a function which is executed for an invalid
	// token.
	JWTErrorHandler func(err error)

	// JWTErrorHandlerWithContext is almost identical to JWTErrorHandler,
	// but it's passed the current context.
	JWTErrorHandlerWithContext func(err error, ctx *http.RequestCtx)

	jwtExtractor func(ctx *http.RequestCtx) ([]byte, error)
)

const (
	SourceCookie = "cookie"
	SourceForm   = "form"
	SourceHeader = "header"
	SourceParam  = "param"
	SourceQuery  = "query"
)

// DefaultJWTConfig is the default JWT auth middleware config.
//nolint: gochecknoglobals
var DefaultJWTConfig = JWTConfig{
	Skipper:       DefaultSkipper,
	SigningMethod: "HS256",
	ContextKey:    "user",
	TokenLookup:   SourceHeader + ":" + http.HeaderAuthorization,
	AuthScheme:    "Bearer",
	Claims:        []jwt.ClaimPair{},
}

var (
	ErrJWTMissing = domain.Error{
		Code:        "invalid_request",
		Description: "missing or malformed jwt",
		URI:         "",
		Frame:       xerrors.Caller(1),
	}

	ErrJWTInvalid = domain.Error{
		Code:        "unauthorized",
		Description: "invalid or expired jwt",
		URI:         "",
		Frame:       xerrors.Caller(1),
	}
)

func JWT(key interface{}) Interceptor {
	return JWTWithConfig(DefaultJWTConfig)
}

//nolint: funlen, gocognit
func JWTWithConfig(config JWTConfig) Interceptor {
	// Defaults
	if config.Skipper == nil {
		config.Skipper = DefaultJWTConfig.Skipper
	}

	if config.SigningKey == nil && len(config.SigningKeys) == 0 && config.ParseTokenFunc == nil {
		panic("jwt middleware requires signing key")
	}

	if config.SigningMethod == "" {
		config.SigningMethod = DefaultJWTConfig.SigningMethod
	}

	if config.ContextKey == "" {
		config.ContextKey = DefaultJWTConfig.ContextKey
	}

	if config.Claims == nil {
		config.Claims = DefaultJWTConfig.Claims
	}

	if config.TokenLookup == "" {
		config.TokenLookup = DefaultJWTConfig.TokenLookup
	}

	if config.AuthScheme == "" {
		config.AuthScheme = DefaultJWTConfig.AuthScheme
	}

	if config.ParseTokenFunc == nil {
		config.ParseTokenFunc = config.defaultParseToken
	}

	extractors := make([]jwtExtractor, 0)
	sources := strings.Split(config.TokenLookup, ",")

	for _, source := range sources {
		parts := strings.Split(source, ":")

		switch parts[0] {
		case SourceQuery:
			extractors = append(extractors, jwtFromQuery(parts[1]))
		case SourceParam:
			extractors = append(extractors, jwtFromParam(parts[1]))
		case SourceCookie:
			extractors = append(extractors, jwtFromCookie(parts[1]))
		case SourceForm:
			extractors = append(extractors, jwtFromForm(parts[1]))
		case SourceHeader:
			extractors = append(extractors, jwtFromHeader(parts[1], config.AuthScheme))
		}
	}

	return func(ctx *http.RequestCtx, next http.RequestHandler) {
		if config.Skipper(ctx) {
			next(ctx)

			return
		}

		if config.BeforeFunc != nil {
			config.BeforeFunc(ctx)
		}

		var (
			auth []byte
			err  error
		)

		for _, extractor := range extractors {
			if auth, err = extractor(ctx); err == nil {
				break
			}
		}

		if err != nil {
			if config.ErrorHandler != nil {
				config.ErrorHandler(err)

				return
			}

			if config.ErrorHandlerWithContext != nil {
				config.ErrorHandlerWithContext(err, ctx)

				return
			}

			ctx.Error(err.Error(), http.StatusInternalServerError)

			return
		}

		if token, err := config.ParseTokenFunc(auth, ctx); err == nil {
			ctx.SetUserValue(config.ContextKey, token)

			if config.SuccessHandler != nil {
				config.SuccessHandler(ctx)
			}

			next(ctx)

			return
		}

		if config.ErrorHandler != nil {
			config.ErrorHandler(err)

			return
		}

		if config.ErrorHandlerWithContext != nil {
			config.ErrorHandlerWithContext(err, ctx)

			return
		}

		ctx.Error(ErrJWTInvalid.Description, http.StatusUnauthorized)
	}
}

func (config *JWTConfig) defaultParseToken(auth []byte, ctx *http.RequestCtx) (interface{}, error) {
	token, err := jwt.Parse(
		auth, jwt.WithVerify(config.SigningMethod, config.SigningKey), jwt.WithValidate(true),
	)
	if err != nil {
		return nil, errors.Wrap(err, "cannot parse JWT token")
	}

	return token, nil
}

func jwtFromHeader(header, authScheme string) jwtExtractor {
	return func(ctx *http.RequestCtx) ([]byte, error) {
		auth := ctx.Request.Header.Peek(header)
		l := len(authScheme)

		if len(auth) > l+1 && bytes.EqualFold(auth[:l], []byte(authScheme)) {
			return auth[l+1:], nil
		}

		return nil, ErrJWTMissing
	}
}

func jwtFromQuery(param string) jwtExtractor {
	return func(ctx *http.RequestCtx) ([]byte, error) {
		if ctx.QueryArgs().Has(param) {
			return nil, ErrJWTMissing
		}

		return ctx.QueryArgs().Peek(param), nil
	}
}

func jwtFromParam(param string) jwtExtractor {
	return func(ctx *http.RequestCtx) ([]byte, error) {
		if !ctx.PostArgs().Has(param) {
			return nil, ErrJWTMissing
		}

		return ctx.PostArgs().Peek(param), nil
	}
}

func jwtFromCookie(name string) jwtExtractor {
	return func(ctx *http.RequestCtx) ([]byte, error) {
		if cookie := ctx.Request.Header.Cookie(name); cookie != nil {
			return cookie, nil
		}

		return nil, ErrJWTMissing
	}
}

func jwtFromForm(name string) jwtExtractor {
	return func(ctx *http.RequestCtx) ([]byte, error) {
		if field := ctx.FormValue(name); field != nil {
			return field, nil
		}

		return nil, ErrJWTMissing
	}
}
