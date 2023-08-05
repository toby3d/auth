package http

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/goccy/go-json"
	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"source.toby3d.me/toby3d/auth/internal/auth"
	"source.toby3d.me/toby3d/auth/internal/client"
	"source.toby3d.me/toby3d/auth/internal/common"
	"source.toby3d.me/toby3d/auth/internal/domain"
	"source.toby3d.me/toby3d/auth/internal/middleware"
	"source.toby3d.me/toby3d/auth/internal/profile"
	"source.toby3d.me/toby3d/auth/internal/urlutil"
	"source.toby3d.me/toby3d/auth/web"
)

type (
	NewHandlerOptions struct {
		Auth     auth.UseCase
		Clients  client.UseCase
		Matcher  language.Matcher
		Profiles profile.UseCase
		Config   domain.Config
	}

	Handler struct {
		clients client.UseCase
		matcher language.Matcher
		useCase auth.UseCase
		config  domain.Config
	}
)

func NewHandler(opts NewHandlerOptions) *Handler {
	return &Handler{
		clients: opts.Clients,
		config:  opts.Config,
		matcher: opts.Matcher,
		useCase: opts.Auth,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	chain := middleware.Chain{
		middleware.CSRFWithConfig(middleware.CSRFConfig{
			Skipper: func(_ http.ResponseWriter, r *http.Request) bool {
				head, _ := urlutil.ShiftPath(r.URL.Path)

				return head == ""
			},
			CookieMaxAge:   0,
			CookieSameSite: http.SameSiteStrictMode,
			ContextKey:     "csrf",
			CookieDomain:   h.config.Server.Domain,
			CookieName:     "__Secure-csrf",
			CookiePath:     "/authorize",
			TokenLookup:    "param:_csrf",
			TokenLength:    0,
			CookieSecure:   true,
			CookieHTTPOnly: true,
		}),
		middleware.BasicAuthWithConfig(middleware.BasicAuthConfig{
			Skipper: func(_ http.ResponseWriter, r *http.Request) bool {
				head, _ := urlutil.ShiftPath(r.URL.Path)

				return r.Method != http.MethodPost || head != "verify" ||
					r.PostFormValue("authorize") == "deny"
			},
			Validator: func(_ http.ResponseWriter, _ *http.Request, login, password string) (bool, error) {
				userMatch := subtle.ConstantTimeCompare([]byte(login),
					[]byte(h.config.IndieAuth.Username))
				passMatch := subtle.ConstantTimeCompare([]byte(password),
					[]byte(h.config.IndieAuth.Password))

				return userMatch == 1 && passMatch == 1, nil
			},
			Realm: "",
		}),
	}

	head, _ := urlutil.ShiftPath(r.URL.Path)

	switch r.Method {
	default:
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	case http.MethodGet, "":
		if head != "" {
			http.NotFound(w, r)

			return
		}

		chain.Handler(h.handleAuthorize).ServeHTTP(w, r)
	case http.MethodPost:
		switch head {
		default:
			http.NotFound(w, r)
		case "":
			chain.Handler(h.handleExchange).ServeHTTP(w, r)
		case "verify":
			chain.Handler(h.handleVerify).ServeHTTP(w, r)
		}
	}
}

func (h *Handler) handleAuthorize(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != "" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)

		return
	}

	w.Header().Set(common.HeaderContentType, common.MIMETextHTMLCharsetUTF8)

	tags, _, _ := language.ParseAcceptLanguage(r.Header.Get(common.HeaderAcceptLanguage))
	tag, _, _ := h.matcher.Match(tags...)
	baseOf := web.BaseOf{
		Config:   &h.config,
		Language: tag,
		Printer:  message.NewPrinter(tag),
	}

	req := NewAuthAuthorizationRequest()
	if err := req.bind(r); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		web.WriteTemplate(w, &web.ErrorPage{
			BaseOf: baseOf,
			Error:  err,
		})

		return
	}

	client, err := h.clients.Discovery(r.Context(), req.ClientID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		web.WriteTemplate(w, &web.ErrorPage{
			BaseOf: baseOf,
			Error:  err,
		})

		return
	}

	if !client.ValidateRedirectURI(req.RedirectURI.URL) {
		w.WriteHeader(http.StatusBadRequest)
		web.WriteTemplate(w, &web.ErrorPage{
			BaseOf: baseOf,
			Error: domain.NewError(domain.ErrorCodeInvalidClient, "requested redirect_uri is not"+
				" registered on client_id side", ""),
		})

		return
	}

	csrf, _ := r.Context().Value(middleware.DefaultCSRFConfig.ContextKey).([]byte)
	web.WriteTemplate(w, &web.AuthorizePage{
		BaseOf:              baseOf,
		CSRF:                csrf,
		Scope:               req.Scope,
		Client:              client,
		Me:                  &req.Me,
		RedirectURI:         &req.RedirectURI,
		CodeChallengeMethod: req.CodeChallengeMethod,
		ResponseType:        req.ResponseType,
		CodeChallenge:       req.CodeChallenge,
		State:               req.State,
		Providers:           make([]*domain.Provider, 0), // TODO(toby3d)
	})
}

func (h *Handler) handleVerify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)

		return
	}

	w.Header().Set(common.HeaderAccessControlAllowOrigin, h.config.Server.Domain)
	w.Header().Set(common.HeaderContentType, common.MIMEApplicationJSONCharsetUTF8)

	encoder := json.NewEncoder(w)

	req := NewAuthVerifyRequest()
	if err := req.bind(r); err != nil {
		w.WriteHeader(http.StatusBadRequest)

		_ = encoder.Encode(err)

		return
	}

	if strings.EqualFold(req.Authorize, "deny") {
		domain.NewError(domain.ErrorCodeAccessDenied, "user deny authorization request", "", req.State).
			SetReirectURI(req.RedirectURI.URL)
		http.Redirect(w, r, req.RedirectURI.String(), http.StatusFound)

		return
	}

	code, err := h.useCase.Generate(r.Context(), auth.GenerateOptions{
		ClientID:            req.ClientID,
		Me:                  req.Me,
		RedirectURI:         req.RedirectURI.URL,
		CodeChallengeMethod: req.CodeChallengeMethod,
		Scope:               req.Scope,
		CodeChallenge:       req.CodeChallenge,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		_ = encoder.Encode(err)

		return
	}

	q := req.RedirectURI.Query()

	for key, val := range map[string]string{
		"code":  code,
		"iss":   h.config.Server.GetRootURL(),
		"state": req.State,
	} {
		q.Set(key, val)
	}

	req.RedirectURI.RawQuery = q.Encode()

	http.Redirect(w, r, req.RedirectURI.String(), http.StatusFound)
}

func (h *Handler) handleExchange(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)

		return
	}

	w.Header().Set(common.HeaderContentType, common.MIMEApplicationJSONCharsetUTF8)

	encoder := json.NewEncoder(w)

	req := new(AuthExchangeRequest)
	if err := req.bind(r); err != nil {
		w.WriteHeader(http.StatusBadRequest)

		_ = encoder.Encode(err)

		return
	}

	me, profile, err := h.useCase.Exchange(r.Context(), auth.ExchangeOptions{
		Code:         req.Code,
		ClientID:     req.ClientID,
		RedirectURI:  req.RedirectURI.URL,
		CodeVerifier: req.CodeVerifier,
	})
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)

		_ = encoder.Encode(err)

		return
	}

	var userInfo *AuthProfileResponse
	if profile != nil {
		userInfo = &AuthProfileResponse{
			Email: profile.GetEmail(),
			Photo: &domain.URL{URL: profile.GetPhoto()},
			URL:   &domain.URL{URL: profile.GetURL()},
			Name:  profile.GetName(),
		}
	}

	_ = encoder.Encode(&AuthExchangeResponse{
		Me:      *me,
		Profile: userInfo,
	})
}
