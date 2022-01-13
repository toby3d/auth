package http

import (
	"strings"

	"github.com/fasthttp/router"
	json "github.com/goccy/go-json"
	http "github.com/valyala/fasthttp"
	"golang.org/x/xerrors"

	"source.toby3d.me/toby3d/form"
	"source.toby3d.me/toby3d/middleware"
	"source.toby3d.me/website/indieauth/internal/common"
	"source.toby3d.me/website/indieauth/internal/domain"
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
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		Scope       string `json:"scope"`
		Me          string `json:"me"`
	}

	//nolint: tagliatelle
	VerificationResponse struct {
		Me       *domain.Me       `json:"me"`
		ClientID *domain.ClientID `json:"client_id"`
		Scope    domain.Scopes    `json:"scope"`
	}

	RevocationResponse struct{}

	RequestHandler struct {
		tokens token.UseCase
		// TODO(toby3d): tickets ticket.UseCase
	}
)

func NewRequestHandler(tokens token.UseCase /*, tickets ticket.UseCase*/) *RequestHandler {
	return &RequestHandler{
		tokens: tokens,
		// tickets: tickets,
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
		encoder.Encode(&domain.Error{
			Code:        "unauthorized_client",
			Description: err.Error(),
			Frame:       xerrors.Caller(1),
		})

		return
	}

	encoder.Encode(&VerificationResponse{
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
			encoder.Encode(domain.Error{
				Code:        "invalid_request",
				Description: err.Error(),
				Frame:       xerrors.Caller(1),
			})

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

	token, err := h.tokens.Exchange(ctx, token.ExchangeOptions{
		ClientID:     req.ClientID,
		RedirectURI:  req.RedirectURI,
		Code:         req.Code,
		CodeVerifier: req.CodeVerifier,
	})
	if err != nil {
		ctx.SetStatusCode(http.StatusBadRequest)
		encoder.Encode(&domain.Error{
			Description: err.Error(),
			Frame:       xerrors.Caller(1),
		})

		return
	}

	encoder.Encode(&ExchangeResponse{
		AccessToken: token.AccessToken,
		TokenType:   "Bearer",
		Scope:       token.Scope.String(),
		Me:          token.Me.String(),
	})
}

func (h *RequestHandler) handleRevoke(ctx *http.RequestCtx) {
	ctx.SetContentType(common.MIMEApplicationJSONCharsetUTF8)
	ctx.SetStatusCode(http.StatusOK)

	encoder := json.NewEncoder(ctx)

	req := new(RevokeRequest)
	if err := req.bind(ctx); err != nil {
		ctx.Error(err.Error(), http.StatusBadRequest)

		return
	}

	if err := h.tokens.Revoke(ctx, req.Token); err != nil {
		ctx.Error(err.Error(), http.StatusBadRequest)

		return
	}

	if err := encoder.Encode(&RevocationResponse{}); err != nil {
		ctx.Error(err.Error(), http.StatusInternalServerError)
	}
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

	/* TODO(toby3d)
	token, err := h.tickets.Redeem(ctx, req.Ticket)
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		encoder.Encode(domain.Error{
			Description: err.Error(),
			Frame:       xerrors.Caller(1),
		})

		return
	}
	*/

	encoder.Encode(ExchangeResponse{})
}

func (r *ExchangeRequest) bind(ctx *http.RequestCtx) error {
	if err := form.Unmarshal(ctx.PostArgs(), r); err != nil {
		return domain.Error{
			Code:        "invalid_request",
			Description: err.Error(),
			Frame:       xerrors.Caller(1),
		}
	}

	return nil
}

func (r *RevokeRequest) bind(ctx *http.RequestCtx) error {
	if err := form.Unmarshal(ctx.PostArgs(), r); err != nil {
		return domain.Error{
			Code:        "invalid_request",
			Description: err.Error(),
			Frame:       xerrors.Caller(1),
		}
	}

	return nil
}

func (r *TicketRequest) bind(ctx *http.RequestCtx) error {
	if err := form.Unmarshal(ctx.PostArgs(), r); err != nil {
		return domain.Error{
			Code:        "invalid_request",
			Description: err.Error(),
			Frame:       xerrors.Caller(1),
		}
	}

	return nil
}
