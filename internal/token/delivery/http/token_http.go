package http

import (
	"errors"
	"strings"

	"github.com/fasthttp/router"
	json "github.com/goccy/go-json"
	http "github.com/valyala/fasthttp"

	"source.toby3d.me/toby3d/form"
	"source.toby3d.me/toby3d/middleware"
	"source.toby3d.me/website/indieauth/internal/common"
	"source.toby3d.me/website/indieauth/internal/domain"
	"source.toby3d.me/website/indieauth/internal/ticket"
	"source.toby3d.me/website/indieauth/internal/token"
)

type (
	ExchangeRequest struct {
		ClientID     *domain.ClientID `form:"client_id"`
		RedirectURI  *domain.URL      `form:"redirect_uri"`
		GrantType    domain.GrantType `form:"grant_type"`
		Code         string           `form:"code"`
		CodeVerifier string           `form:"code_verifier"`
	}

	RevokeRequest struct {
		Action domain.Action `form:"action"`
		Token  string        `form:"token"`
	}

	TicketRequest struct {
		Action domain.Action `form:"action"`
		Ticket string        `form:"ticket"`
	}

	//nolint: tagliatelle
	ExchangeResponse struct {
		AccessToken string           `json:"access_token"`
		TokenType   string           `json:"token_type"`
		Scope       string           `json:"scope"`
		Me          string           `json:"me"`
		Profile     *ProfileResponse `json:"profile,omitempty"`
	}

	ProfileResponse struct {
		Name  string        `json:"name,omitempty"`
		URL   *domain.URL   `json:"url,omitempty"`
		Photo *domain.URL   `json:"photo,omitempty"`
		Email *domain.Email `json:"email,omitempty"`
	}

	//nolint: tagliatelle
	VerificationResponse struct {
		Me       *domain.Me       `json:"me"`
		ClientID *domain.ClientID `json:"client_id"`
		Scope    domain.Scopes    `json:"scope"`
	}

	RevocationResponse struct{}

	RequestHandler struct {
		tokens  token.UseCase
		tickets ticket.UseCase
	}
)

func NewRequestHandler(tokens token.UseCase, tickets ticket.UseCase) *RequestHandler {
	return &RequestHandler{
		tokens:  tokens,
		tickets: tickets,
	}
}

func (h *RequestHandler) Register(r *router.Router) {
	chain := middleware.Chain{
		middleware.LogFmt(),
	}

	r.GET("/token", chain.RequestHandler(h.handleValidate))
	r.POST("/token", chain.RequestHandler(h.handleAction))
}

func (h *RequestHandler) handleValidate(ctx *http.RequestCtx) {
	ctx.SetContentType(common.MIMEApplicationJSONCharsetUTF8)
	ctx.SetStatusCode(http.StatusOK)

	encoder := json.NewEncoder(ctx)

	t, err := h.tokens.Verify(ctx, strings.TrimPrefix(string(ctx.Request.Header.Peek(http.HeaderAuthorization)),
		"Bearer "))
	if err != nil || t == nil {
		ctx.SetStatusCode(http.StatusUnauthorized)
		encoder.Encode(domain.NewError(
			domain.ErrorCodeUnauthorizedClient,
			err.Error(),
			"https://indieauth.net/source/#access-token-verification",
		))

		return
	}

	_ = encoder.Encode(&VerificationResponse{
		ClientID: t.ClientID,
		Me:       t.Me,
		Scope:    t.Scope,
	})
}

func (h *RequestHandler) handleAction(ctx *http.RequestCtx) {
	ctx.SetContentType(common.MIMEApplicationJSONCharsetUTF8)

	encoder := json.NewEncoder(ctx)

	switch {
	case ctx.PostArgs().Has("grant_type"):
		h.handleExchange(ctx)
	case ctx.PostArgs().Has("action"):
		action, err := domain.ParseAction(string(ctx.PostArgs().Peek("action")))
		if err != nil {
			ctx.SetStatusCode(http.StatusBadRequest)
			encoder.Encode(domain.NewError(
				domain.ErrorCodeInvalidRequest,
				err.Error(),
				"",
			))

			return
		}

		switch action {
		case domain.ActionRevoke:
			h.handleRevoke(ctx)
		case domain.ActionTicket:
			h.handleTicket(ctx)
		}
	}
}

func (h *RequestHandler) handleExchange(ctx *http.RequestCtx) {
	ctx.SetContentType(common.MIMEApplicationJSONCharsetUTF8)

	encoder := json.NewEncoder(ctx)

	req := new(ExchangeRequest)
	if err := req.bind(ctx); err != nil {
		ctx.SetStatusCode(http.StatusBadRequest)
		encoder.Encode(err)

		return
	}

	token, profile, err := h.tokens.Exchange(ctx, token.ExchangeOptions{
		ClientID:     req.ClientID,
		RedirectURI:  req.RedirectURI,
		Code:         req.Code,
		CodeVerifier: req.CodeVerifier,
	})
	if err != nil {
		ctx.SetStatusCode(http.StatusBadRequest)
		encoder.Encode(domain.NewError(
			domain.ErrorCodeInvalidRequest,
			err.Error(),
			"https://indieauth.net/source/#request",
		))

		return
	}

	resp := &ExchangeResponse{
		AccessToken: token.AccessToken,
		TokenType:   "Bearer",
		Scope:       token.Scope.String(),
		Me:          token.Me.String(),
	}

	if profile == nil {
		encoder.Encode(resp)

		return
	}

	resp.Profile = new(ProfileResponse)

	if len(profile.Name) > 0 {
		resp.Profile.Name = profile.Name[0]
	}

	if len(profile.URL) > 0 {
		resp.Profile.URL = profile.URL[0]
	}

	if len(profile.Photo) > 0 {
		resp.Profile.Photo = profile.Photo[0]
	}

	if len(profile.Email) > 0 {
		resp.Profile.Email = profile.Email[0]
	}

	_ = encoder.Encode(resp)
}

func (h *RequestHandler) handleRevoke(ctx *http.RequestCtx) {
	ctx.SetContentType(common.MIMEApplicationJSONCharsetUTF8)
	ctx.SetStatusCode(http.StatusOK)

	encoder := json.NewEncoder(ctx)

	req := new(RevokeRequest)
	if err := req.bind(ctx); err != nil {
		ctx.SetStatusCode(http.StatusBadRequest)
		encoder.Encode(err)

		return
	}

	if err := h.tokens.Revoke(ctx, req.Token); err != nil {
		ctx.SetStatusCode(http.StatusBadRequest)
		encoder.Encode(domain.NewError(
			domain.ErrorCodeInvalidRequest,
			err.Error(),
			"",
		))

		return
	}

	_ = encoder.Encode(&RevocationResponse{})
}

func (h *RequestHandler) handleTicket(ctx *http.RequestCtx) {
	ctx.SetContentType(common.MIMEApplicationJSONCharsetUTF8)
	ctx.SetStatusCode(http.StatusOK)

	encoder := json.NewEncoder(ctx)

	req := new(TicketRequest)
	if err := req.bind(ctx); err != nil {
		ctx.SetStatusCode(http.StatusBadRequest)
		encoder.Encode(err)

		return
	}

	t, err := h.tickets.Exchange(ctx, req.Ticket)
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		encoder.Encode(domain.NewError(
			domain.ErrorCodeInvalidRequest,
			err.Error(),
			"https://indieauth.net/source/#request",
		))

		return
	}

	encoder.Encode(ExchangeResponse{
		AccessToken: t.AccessToken,
		TokenType:   "Bearer",
		Scope:       t.Scope.String(),
		Me:          t.Me.String(),
	})
}

func (r *ExchangeRequest) bind(ctx *http.RequestCtx) error {
	indieAuthError := new(domain.Error)
	if err := form.Unmarshal(ctx.PostArgs(), r); err != nil {
		if errors.As(err, indieAuthError) {
			return indieAuthError
		}

		return domain.NewError(
			domain.ErrorCodeInvalidRequest,
			err.Error(),
			"https://indieauth.net/source/#request",
		)
	}

	return nil
}

func (r *RevokeRequest) bind(ctx *http.RequestCtx) error {
	indieAuthError := new(domain.Error)
	if err := form.Unmarshal(ctx.PostArgs(), r); err != nil {
		if errors.As(err, indieAuthError) {
			return indieAuthError
		}

		return domain.NewError(
			domain.ErrorCodeInvalidRequest,
			err.Error(),
			"https://indieauth.net/source/#request",
		)
	}

	return nil
}

func (r *TicketRequest) bind(ctx *http.RequestCtx) error {
	indieAuthError := new(domain.Error)
	if err := form.Unmarshal(ctx.PostArgs(), r); err != nil {
		if errors.As(err, indieAuthError) {
			return indieAuthError
		}

		return domain.NewError(
			domain.ErrorCodeInvalidRequest,
			err.Error(),
			"https://indieauth.net/source/#request",
		)
	}

	return nil
}
