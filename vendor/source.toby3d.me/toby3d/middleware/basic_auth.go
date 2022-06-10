package middleware

import (
	"encoding/base64"
	"strconv"
	"strings"

	http "github.com/valyala/fasthttp"
)

type (
	BasicAuthConfig struct {
		Skipper   Skipper
		Validator BasicAuthValidator
		Realm     string
	}

	BasicAuthValidator func(ctx *http.RequestCtx, login, password string) (bool, error)
)

const DefaultRealm string = "Restricted"

const basic string = "basic"

//nolint: gochecknoglobals
var DefaultBasicAuthConfig = BasicAuthConfig{
	Skipper: DefaultSkipper,
	Realm:   DefaultRealm,
}

func BasicAuth(validator BasicAuthValidator) Interceptor {
	cfg := DefaultBasicAuthConfig
	cfg.Validator = validator

	return BasicAuthWithConfig(cfg)
}

func BasicAuthWithConfig(config BasicAuthConfig) Interceptor {
	// Defaults
	if config.Validator == nil {
		panic("middleware: basic-auth middleware requires a validator function")
	}

	if config.Skipper == nil {
		config.Skipper = DefaultBasicAuthConfig.Skipper
	}

	if config.Realm == "" {
		config.Realm = DefaultRealm
	}

	return func(ctx *http.RequestCtx, next http.RequestHandler) {
		if config.Skipper(ctx) {
			next(ctx)

			return
		}

		auth := ctx.Request.Header.Peek(http.HeaderAuthorization)
		l := len(basic)

		if len(auth) > l+1 && strings.EqualFold(string(auth[:l]), basic) {
			b, err := base64.StdEncoding.DecodeString(string(auth[l+1:]))
			if err != nil {
				ctx.Error(err.Error(), http.StatusBadRequest)

				return
			}

			cred := string(b)
			for i := 0; i < len(cred); i++ {
				if cred[i] == ':' {
					// NOTE(toby3d): verify credentials.
					valid, err := config.Validator(ctx, cred[:i], cred[i+1:])
					if err != nil {
						ctx.Error(err.Error(), http.StatusBadRequest)

						return
					}

					if valid {
						next(ctx)

						return
					}

					break
				}
			}
		}

		realm := DefaultRealm
		if config.Realm != DefaultRealm {
			realm = strconv.Quote(config.Realm)
		}

		// NOTE(toby3d): require 401 status for login pop-up.
		ctx.SetStatusCode(http.StatusUnauthorized)
		ctx.Response.Header.Set(http.HeaderWWWAuthenticate, basic+" realm="+realm)
	}
}
