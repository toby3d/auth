package middleware

import (
	"encoding/base64"
	"net/http"
	"strconv"
	"strings"

	"source.toby3d.me/toby3d/auth/internal/common"
)

type (
	BasicAuthConfig struct {
		Skipper   Skipper
		Validator BasicAuthValidator
		Realm     string
	}

	BasicAuthValidator func(w http.ResponseWriter, r *http.Request, login, password string) (bool, error)
)

const DefaultRealm string = "Restricted"

const basic string = "basic"

// nolint: gochecknoglobals
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

	return func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		if config.Skipper(w, r) {
			next(w, r)

			return
		}

		auth := r.Header.Get(common.HeaderAuthorization)
		l := len(basic)

		if len(auth) > l+1 && strings.EqualFold(string(auth[:l]), basic) {
			b, err := base64.StdEncoding.DecodeString(string(auth[l+1:]))
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)

				return
			}

			cred := string(b)
			for i := 0; i < len(cred); i++ {
				if cred[i] == ':' {
					// NOTE(toby3d): verify credentials.
					valid, err := config.Validator(w, r, cred[:i], cred[i+1:])
					if err != nil {
						http.Error(w, err.Error(), http.StatusBadRequest)

						return
					}

					if valid {
						next(w, r)

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
		w.Header().Set(common.HeaderWWWAuthenticate, basic+" realm="+realm)
		w.WriteHeader(http.StatusUnauthorized)
	}
}
