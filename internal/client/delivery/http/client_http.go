package http

import (
	"net/http"
	"strings"

	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"source.toby3d.me/toby3d/auth/internal/common"
	"source.toby3d.me/toby3d/auth/internal/domain"
	"source.toby3d.me/toby3d/auth/internal/token"
	"source.toby3d.me/toby3d/auth/internal/urlutil"
	"source.toby3d.me/toby3d/auth/web"
)

type (
	NewHandlerOptions struct {
		Matcher language.Matcher
		Tokens  token.UseCase
		Client  domain.Client
		Config  domain.Config
	}

	Handler struct {
		matcher language.Matcher
		tokens  token.UseCase
		client  domain.Client
		config  domain.Config
	}
)

func NewHandler(opts NewHandlerOptions) *Handler {
	return &Handler{
		client:  opts.Client,
		config:  opts.Config,
		matcher: opts.Matcher,
		tokens:  opts.Tokens,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "" && r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)

		return
	}

	var head string
	head, r.URL.Path = urlutil.ShiftPath(r.URL.Path)

	switch head {
	default:
		http.NotFound(w, r)
	case "":
		h.handleRender(w, r)
	case "callback":
		h.handleCallback(w, r)
	}
}

func (h *Handler) handleRender(w http.ResponseWriter, r *http.Request) {
	if r.Method != "" && r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)

		return
	}

	redirect := make([]string, len(h.client.RedirectURI))

	for i := range h.client.RedirectURI {
		redirect[i] = h.client.RedirectURI[i].String()
	}

	w.Header().Set(common.HeaderLink, `<`+strings.Join(redirect, `>; rel="redirect_uri", `)+`>; rel="redirect_uri"`)

	tags, _, _ := language.ParseAcceptLanguage(r.Header.Get(common.HeaderAcceptLanguage))
	tag, _, _ := h.matcher.Match(tags...)

	// TODO(toby3d): generate and store PKCE

	w.Header().Set(common.HeaderContentType, common.MIMETextHTMLCharsetUTF8)
	web.WriteTemplate(w, &web.HomePage{
		BaseOf: web.BaseOf{
			Config:   &h.config,
			Language: tag,
			Printer:  message.NewPrinter(tag),
		},
		Client: &h.client,
		State:  "hackme", // TODO(toby3d): generate and store state
	})
}

//nolint:unlen
func (h *Handler) handleCallback(w http.ResponseWriter, r *http.Request) {
	if r.Method != "" && r.Method != http.MethodGet {
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

	req := new(ClientCallbackRequest)
	if err := req.bind(r); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		web.WriteTemplate(w, &web.ErrorPage{
			BaseOf: baseOf,
			Error:  err,
		})

		return
	}

	if req.Error != domain.ErrorCodeUnd {
		w.WriteHeader(http.StatusUnauthorized)
		web.WriteTemplate(w, &web.ErrorPage{
			BaseOf: baseOf,
			Error: domain.NewError(
				domain.ErrorCodeAccessDenied,
				req.ErrorDescription,
				"",
				req.State,
			),
		})

		return
	}

	// TODO(toby3d): load and check state

	if req.Iss.String() != h.client.ID.String() {
		w.WriteHeader(http.StatusBadRequest)
		web.WriteTemplate(w, &web.ErrorPage{
			BaseOf: baseOf,
			Error: domain.NewError(
				domain.ErrorCodeInvalidClient,
				"iss does not match client_id",
				"https://indieauth.net/source/#authorization-response",
				req.State,
			),
		})

		return
	}

	token, _, err := h.tokens.Exchange(r.Context(), token.ExchangeOptions{
		ClientID:     h.client.ID,
		RedirectURI:  h.client.RedirectURI[0],
		Code:         req.Code,
		CodeVerifier: "", // TODO(toby3d): validate PKCE here
	})
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		web.WriteTemplate(w, &web.ErrorPage{
			BaseOf: baseOf,
			Error:  err,
		})

		return
	}

	w.Header().Set(common.HeaderContentType, common.MIMETextHTMLCharsetUTF8)
	web.WriteTemplate(w, &web.CallbackPage{
		BaseOf: baseOf,
		Token:  token,
	})
}
