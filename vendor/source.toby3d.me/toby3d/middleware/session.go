package middleware

import (
	"fmt"

	"github.com/fasthttp/session/v2"
	http "github.com/valyala/fasthttp"
)

type SessionConfig struct {
	Skipper Skipper
	Session *session.Session
}

// DefaultSessionConfig is the default Session middleware config.
var DefaultSessionConfig = SessionConfig{
	Skipper: DefaultSkipper,
}

var key string = "_session_store"

func Session(s *session.Session) Interceptor {
	c := DefaultSessionConfig
	c.Session = s

	return SessionWithConfig(c)
}

func SessionWithConfig(config SessionConfig) Interceptor {
	if config.Skipper == nil {
		config.Skipper = DefaultSessionConfig.Skipper
	}
	if config.Session == nil {
		panic("session: session middleware requires session")
	}

	return func(ctx *http.RequestCtx, next http.RequestHandler) {
		if config.Skipper(ctx) {
			next(ctx)

			return
		}

		store, err := config.Session.Get(ctx)
		if err != nil {
			ctx.Error(err.Error(), http.StatusInternalServerError)

			return
		}

		defer func(s *session.Session, ctx *http.RequestCtx, store *session.Store) {
			if err := s.Save(ctx, store); err != nil {
				ctx.Error(err.Error(), http.StatusInternalServerError)
			}
		}(config.Session, ctx, store)

		ctx.SetUserValue(key, store)
		next(ctx)
	}
}

// SessionGet returns a named session.
func SessionGet(ctx *http.RequestCtx) (*session.Store, error) {
	store, ok := ctx.UserValue(key).(*session.Store)
	if !ok {
		return nil, fmt.Errorf("%s session store not found", key)
	}

	return store, nil
}
