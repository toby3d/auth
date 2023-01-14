package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwt"

	"source.toby3d.me/toby3d/auth/internal/common"
)

type (
	// JWTConfig defines the config for JWT middleware.
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
		// KeyFunc, SigningKeys and SigningKey.
		//
		// Required if neither user-defined KeyFunc nor SigningKeys is
		// provided.
		SigningKey any

		// Map of signing keys to validate token with kid field usage.
		// This is one of the three options to provide a token
		// validation key. The order of precedence is a user-defined
		// KeyFunc, SigningKeys and SigningKey.
		//
		// Required if neither user-defined KeyFunc nor SigningKey is
		// provided.
		SigningKeys map[string]any

		// Signing method used to check the token's signing algorithm.
		//
		// Optional. Default value HS256.
		SigningMethod jwa.SignatureAlgorithm

		// Context key to store user information from the token into
		// context.
		//
		// Optional. Default value "user".
		ContextKey string

		// Claims are extendable claims data defining token content.
		// Used by default ParseTokenFunc implementation. Not used if
		// custom ParseTokenFunc is set.
		//
		// Optional. Default value []jwt.ClaimPair
		Claims []jwt.ClaimPair

		// TokenLookup is a string in the form of "<source>:<name>" or
		// "<source>:<name>,<source>:<name>" that is used to extract
		// token from the request.
		// Optional. Default value "header:Authorization".
		// Possible values:
		// - "header:<name>"
		// - "query:<name>"
		// - "cookie:<name>"
		// - "form:<name>"
		// Multiply sources example:
		// - "header: Authorization,cookie: myowncookie"
		TokenLookup string

		// TokenLookupFuncs defines a list of user-defined functions
		// that extract JWT token from the given context.
		// This is one of the two options to provide a token extractor.
		// The order of precedence is user-defined TokenLookupFuncs, and
		// TokenLookup.
		// You can also provide both if you want.
		TokenLookupFuncs []ValuesExtractor

		// AuthScheme to be used in the Authorization header.
		//
		// Optional. Default value "Bearer".
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
		ParseTokenFunc func(auth []byte, w http.ResponseWriter, r *http.Request) (any, error)

		// ContinueOnIgnoredError allows the next middleware/handler to
		// be called when ErrorHandlerWithContext decides to ignore the
		// error (by returning `nil`). This is useful when parts of your
		// site/api allow public access and some authorized routes
		// provide extra functionality. In that case you can use
		// ErrorHandlerWithContext to set a default public JWT token
		// value in the request context and continue. Some logic down
		// the remaining execution chain needs to check that (public)
		// token value then.
		ContinueOnIgnoredError bool
	}

	// JWTSuccessHandler defines a function which is executed for a valid
	// token.
	JWTSuccessHandler func(w http.ResponseWriter, r *http.Request)

	// JWTErrorHandler defines a function which is executed for an invalid
	// token.
	JWTErrorHandler func(err error)

	// JWTErrorHandlerWithContext is almost identical to JWTErrorHandler,
	// but it's passed the current context.
	JWTErrorHandlerWithContext func(err error, w http.ResponseWriter, r *http.Request) error

	jwtExtractor func(w http.ResponseWriter, r *http.Request) ([]byte, error)
)

// DefaultJWTConfig is the default JWT auth middleware config.
//
//nolint:gochecknoglobals
var DefaultJWTConfig = JWTConfig{
	Skipper:       DefaultSkipper,
	SigningMethod: "HS256",
	ContextKey:    "user",
	TokenLookup:   "header:" + common.HeaderAuthorization,
	AuthScheme:    "Bearer",
	Claims:        []jwt.ClaimPair{},
}

// Possible token errors.
var (
	ErrJWTMissing = errors.New("jwt: missing or malformed jwt")
	ErrJWTInvalid = errors.New("jwt: invalid or expired jwt")
)

// JWT returns a JSON Web Token (JWT) auth middleware.
//
// * For valid token, it sets the user in context and calls next handler.
// * For invalid token, it returns "401 - Unauthorized" error.
// * For missing token, it returns "400 - Bad Request" error.
func JWT(key any) Interceptor {
	config := DefaultJWTConfig
	config.SigningKey = key

	return JWTWithConfig(config)
}

// JWTWithConfig returns a JWT auth middleware with config.
//
//nolint:funlen,gocognit
func JWTWithConfig(config JWTConfig) Interceptor {
	// Defaults
	if config.Skipper == nil {
		config.Skipper = DefaultJWTConfig.Skipper
	}

	if config.SigningKey == nil && len(config.SigningKeys) == 0 && config.ParseTokenFunc == nil {
		panic("jwt: middleware requires signing key")
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

	extractors, err := createExtractors(config.TokenLookup, config.AuthScheme)
	if err != nil {
		panic("middleware: jwt: " + err.Error())
	}

	if len(config.TokenLookupFuncs) > 0 {
		extractors = append(config.TokenLookupFuncs, extractors...)
	}

	return func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		if config.Skipper(w, r) {
			next(w, r)

			return
		}

		if config.BeforeFunc != nil {
			config.BeforeFunc(w, r)
		}

		var lastExtractorErr, lastTokenErr error

		for _, extractor := range extractors {
			auths, err := extractor(w, r)
			if err != nil {
				lastExtractorErr = ErrJWTMissing

				continue
			}

			for _, auth := range auths {
				token, err := config.ParseTokenFunc([]byte(auth), w, r)
				if err != nil {
					lastTokenErr = err

					continue
				}

				r = r.WithContext(context.WithValue(r.Context(), config.ContextKey, token))

				if config.SuccessHandler != nil {
					config.SuccessHandler(w, r)
				}

				next(w, r)

				return
			}
		}

		err := lastTokenErr
		if err == nil {
			err = lastExtractorErr
		}

		if config.ErrorHandler != nil {
			config.ErrorHandler(err)

			return
		}

		if config.ErrorHandlerWithContext != nil {
			tmpErr := config.ErrorHandlerWithContext(err, w, r)
			if config.ContinueOnIgnoredError && tmpErr == nil {
				next(w, r)
			} else {
				http.Error(w, tmpErr.Error(), http.StatusUnauthorized)
			}

			return
		}

		if lastTokenErr != nil {
			http.Error(w, lastTokenErr.Error(), http.StatusUnauthorized)

			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (config *JWTConfig) defaultParseToken(auth []byte, w http.ResponseWriter, r *http.Request) (any, error) {
	token, err := jwt.Parse(auth, jwt.WithKey(config.SigningMethod, config.SigningKey), jwt.WithVerify(true))
	if err != nil {
		return nil, fmt.Errorf("cannot parse JWT token: %w", err)
	}

	return token, nil
}
