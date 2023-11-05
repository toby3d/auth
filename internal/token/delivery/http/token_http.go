package http

import (
	"net/http"

	"github.com/goccy/go-json"
	"github.com/lestrrat-go/jwx/v2/jwa"

	"source.toby3d.me/toby3d/auth/internal/common"
	"source.toby3d.me/toby3d/auth/internal/domain"
	"source.toby3d.me/toby3d/auth/internal/middleware"
	"source.toby3d.me/toby3d/auth/internal/token"
	"source.toby3d.me/toby3d/auth/internal/urlutil"
)

type Handler struct {
	config domain.Config
	tokens token.UseCase
}

func NewHandler(tokens token.UseCase, config domain.Config) *Handler {
	return &Handler{
		config: config,
		tokens: tokens,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	chain := middleware.Chain{
		//nolint:exhaustivestruct
		middleware.JWTWithConfig(middleware.JWTConfig{
			Skipper: func(_ http.ResponseWriter, r *http.Request) bool {
				head, _ := urlutil.ShiftPath(r.URL.Path)

				return head == "token"
			},
			SigningKey:    []byte(h.config.JWT.Secret),
			SigningMethod: jwa.SignatureAlgorithm(h.config.JWT.Algorithm),
			ContextKey:    "token",
			TokenLookup:   "form:token," + "header:" + common.HeaderAuthorization + ":Bearer ",
			AuthScheme:    "Bearer",
		}),
	}

	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)

		return
	}

	var head string
	head, r.URL.Path = urlutil.ShiftPath(r.URL.Path)

	switch head {
	default:
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	case "token":
		chain.Handler(h.handleAction).ServeHTTP(w, r)
	case "introspect":
		chain.Handler(h.handleIntrospect).ServeHTTP(w, r)
	case "revocation":
		chain.Handler(h.handleRevokation).ServeHTTP(w, r)
	}
}

func (h *Handler) handleIntrospect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)

		return
	}

	w.Header().Set(common.HeaderContentType, common.MIMEApplicationJSONCharsetUTF8)

	encoder := json.NewEncoder(w)

	req := new(TokenIntrospectRequest)
	if err := req.bind(r); err != nil {
		_ = encoder.Encode(err)

		w.WriteHeader(http.StatusBadRequest)

		return
	}

	tkn, _, err := h.tokens.Verify(r.Context(), req.Token)
	if err != nil || tkn == nil {
		// WARN(toby3d): If the token is not valid, the endpoint still
		// MUST return a 200 Response.
		_ = encoder.Encode(&TokenInvalidIntrospectResponse{Active: false})

		w.WriteHeader(http.StatusOK)

		return
	}

	_ = encoder.Encode(&TokenIntrospectResponse{
		Active:   true,
		ClientID: tkn.ClientID.String(),
		Exp:      tkn.Expiry.Unix(),
		Iat:      tkn.CreatedAt.Unix(),
		Me:       tkn.Me.String(),
		Scope:    tkn.Scope.String(),
	})

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) handleAction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)

		return
	}

	w.Header().Set(common.HeaderContentType, common.MIMEApplicationJSONCharsetUTF8)

	encoder := json.NewEncoder(w)

	switch {
	case r.PostForm.Has("grant_type"):
		h.handleExchange(w, r)
	case r.PostForm.Has("action"):
		if err := r.ParseForm(); err != nil {
			w.WriteHeader(http.StatusBadRequest)

			_ = encoder.Encode(domain.NewError(domain.ErrorCodeInvalidRequest, err.Error(), ""))

			return
		}

		action, err := domain.ParseAction(r.PostForm.Get("action"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)

			_ = encoder.Encode(domain.NewError(domain.ErrorCodeInvalidRequest, err.Error(), ""))

			return
		}

		switch action {
		case domain.ActionRevoke:
			h.handleRevokation(w, r)
		}
	}
}

func (h *Handler) handleExchange(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)

		return
	}

	w.Header().Set(common.HeaderContentType, common.MIMEApplicationJSONCharsetUTF8)

	encoder := json.NewEncoder(w)

	req := new(TokenExchangeRequest)
	if err := req.bind(r); err != nil {
		w.WriteHeader(http.StatusBadRequest)

		_ = encoder.Encode(err)

		return
	}

	token, profile, err := h.tokens.Exchange(r.Context(), token.ExchangeOptions{
		ClientID:     req.ClientID,
		RedirectURI:  req.RedirectURI.URL,
		Code:         req.Code,
		CodeVerifier: req.CodeVerifier,
	})
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)

		_ = encoder.Encode(domain.NewError(domain.ErrorCodeInvalidRequest, err.Error(),
			"https://indieauth.net/source/#request"))

		return
	}

	resp := &TokenExchangeResponse{
		AccessToken:  token.AccessToken,
		ExpiresIn:    token.Expiry.Unix(),
		Me:           token.Me.String(),
		Profile:      NewTokenProfileResponse(profile),
		RefreshToken: "", // TODO(toby3d)
	}

	_ = encoder.Encode(resp)

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) handleRevokation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)

		return
	}

	w.Header().Set(common.HeaderContentType, common.MIMEApplicationJSONCharsetUTF8)

	encoder := json.NewEncoder(w)

	req := NewTokenRevocationRequest()
	if err := req.bind(r); err != nil {
		w.WriteHeader(http.StatusBadRequest)

		_ = encoder.Encode(err)

		return
	}

	if err := h.tokens.Revoke(r.Context(), req.Token); err != nil {
		w.WriteHeader(http.StatusBadRequest)

		_ = encoder.Encode(domain.NewError(domain.ErrorCodeInvalidRequest, err.Error(), ""))

		return
	}

	_ = encoder.Encode(&TokenRevocationResponse{})

	w.WriteHeader(http.StatusOK)
}
