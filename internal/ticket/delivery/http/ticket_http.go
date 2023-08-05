package http

import (
	"fmt"
	"net/http"

	"github.com/goccy/go-json"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"source.toby3d.me/toby3d/auth/internal/common"
	"source.toby3d.me/toby3d/auth/internal/domain"
	"source.toby3d.me/toby3d/auth/internal/middleware"
	"source.toby3d.me/toby3d/auth/internal/random"
	"source.toby3d.me/toby3d/auth/internal/ticket"
	"source.toby3d.me/toby3d/auth/internal/urlutil"
	"source.toby3d.me/toby3d/auth/web"
)

type Handler struct {
	matcher language.Matcher
	tickets ticket.UseCase
	config  domain.Config
}

func NewHandler(tickets ticket.UseCase, matcher language.Matcher, config domain.Config) *Handler {
	return &Handler{
		config:  config,
		matcher: matcher,
		tickets: tickets,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	//nolint:exhaustivestruct
	chain := middleware.Chain{
		middleware.CSRFWithConfig(middleware.CSRFConfig{
			Skipper: func(_ http.ResponseWriter, r *http.Request) bool {
				head, _ := urlutil.ShiftPath(r.URL.Path)

				return r.Method == http.MethodPost && head == "ticket"
			},
			CookieMaxAge:   0,
			CookieSameSite: http.SameSiteStrictMode,
			ContextKey:     "csrf",
			CookieDomain:   h.config.Server.Domain,
			CookieName:     "__Secure-csrf",
			CookiePath:     "/ticket",
			TokenLookup:    "form:_csrf",
			TokenLength:    0,
			CookieSecure:   true,
			CookieHTTPOnly: true,
		}),
		middleware.JWTWithConfig(middleware.JWTConfig{
			AuthScheme:              "Bearer",
			BeforeFunc:              nil,
			Claims:                  nil,
			ContextKey:              "token",
			ErrorHandler:            nil,
			ErrorHandlerWithContext: nil,
			ParseTokenFunc:          nil,
			SigningKey:              []byte(h.config.JWT.Secret),
			SigningKeys:             nil,
			SigningMethod:           jwa.SignatureAlgorithm(h.config.JWT.Algorithm),
			Skipper:                 middleware.DefaultSkipper,
			SuccessHandler:          nil,
			TokenLookup: "header:" + common.HeaderAuthorization +
				",cookie:__Secure-auth-token",
		}),
	}

	chain.Handler(h.handleFunc).ServeHTTP(w, r)
}

func (h *Handler) handleFunc(w http.ResponseWriter, r *http.Request) {
	var head string
	head, r.URL.Path = urlutil.ShiftPath(r.URL.Path)

	switch r.Method {
	default:
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	case "", http.MethodGet:
		if head != "" {
			http.NotFound(w, r)

			return
		}

		h.handleRender(w, r)
	case http.MethodPost:
		switch head {
		default:
			http.NotFound(w, r)
		case "":
			h.handleRedeem(w, r)
		case "send":
			h.handleSend(w, r)
		}
	}
}

func (h *Handler) handleRender(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(common.HeaderContentType, common.MIMETextHTMLCharsetUTF8)

	tags, _, _ := language.ParseAcceptLanguage(r.Header.Get(common.HeaderAcceptLanguage))
	tag, _, _ := h.matcher.Match(tags...)
	baseOf := web.BaseOf{
		Config:   &h.config,
		Language: tag,
		Printer:  message.NewPrinter(tag),
	}

	csrf, _ := r.Context().Value("csrf").([]byte)
	web.WriteTemplate(w, &web.TicketPage{
		BaseOf: baseOf,
		CSRF:   csrf,
	})
}

func (h *Handler) handleSend(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(common.HeaderAccessControlAllowOrigin, h.config.Server.Domain)
	w.Header().Set(common.HeaderContentType, common.MIMEApplicationJSONCharsetUTF8)

	encoder := json.NewEncoder(w)

	req := new(TicketGenerateRequest)
	if err := req.bind(r); err != nil {
		w.WriteHeader(http.StatusBadRequest)

		_ = encoder.Encode(err)

		return
	}

	ticket := &domain.Ticket{
		Ticket:   "",
		Resource: req.Resource.URL,
		Subject:  &req.Subject,
	}

	var err error
	if ticket.Ticket, err = random.String(h.config.TicketAuth.Length); err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		_ = encoder.Encode(domain.NewError(domain.ErrorCodeServerError, err.Error(), ""))

		return
	}

	if err = h.tickets.Generate(r.Context(), *ticket); err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		_ = encoder.Encode(domain.NewError(domain.ErrorCodeServerError, err.Error(), ""))

		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) handleRedeem(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(common.HeaderContentType, common.MIMEApplicationJSONCharsetUTF8)

	encoder := json.NewEncoder(w)

	req := new(TicketExchangeRequest)
	if err := req.bind(r); err != nil {
		w.WriteHeader(http.StatusBadRequest)

		_ = encoder.Encode(err)

		return
	}

	token, err := h.tickets.Redeem(r.Context(), domain.Ticket{
		Ticket:   req.Ticket,
		Resource: req.Resource.URL,
		Subject:  &req.Subject,
	})
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)

		_ = encoder.Encode(domain.NewError(domain.ErrorCodeServerError, err.Error(), ""))

		return
	}

	// TODO(toby3d): print the result as part of the debugging. Instead, we
	// need to send or save the token to the recipient for later use.
	fmt.Fprintf(w, `{
		"access_token": "%s",
		"token_type": "Bearer",
		"scope": "%s",
		"me": "%s"
	}`, token.AccessToken, token.Scope.String(), token.Me.String())
	w.WriteHeader(http.StatusOK)
}
