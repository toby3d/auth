package http

import (
	"strings"

	"github.com/fasthttp/router"
	json "github.com/goccy/go-json"
	http "github.com/valyala/fasthttp"
	"gitlab.com/toby3d/indieauth/internal/model"
	"gitlab.com/toby3d/indieauth/internal/token"
)

type (
	Handler struct {
		useCase token.UseCase
	}

	ExchangeRequest struct {
		GrantType    string
		Code         string
		ClientID     string
		RedirectURI  string
		CodeVerifier string
	}

	RevokeRequest struct {
		Action string
		Token  string
	}

	ExchangeResponse struct {
		AccessToken string `json:"access_token"`
		Me          string `json:"me"`
		Scope       string `json:"scope"`
		TokenType   string `json:"token_type"`
	}

	VerificationResponse struct {
		Me       string `json:"me"`
		ClientID string `json:"client_id"`
		Scope    string `json:"scope"`
	}
)

func NewTokenHandler(useCase token.UseCase) *Handler {
	return &Handler{
		useCase: useCase,
	}
}

func (h *Handler) Register(r *router.Router) {
	r.GET("/token", h.Verification)
	r.POST("/token", h.Update)
}

func (h *Handler) Verification(ctx *http.RequestCtx) {
	encoder := json.NewEncoder(ctx)

	token, err := h.useCase.Verify(ctx,
		strings.TrimPrefix(string(ctx.Request.Header.Peek(http.HeaderAuthorization)), "Bearer "),
	)
	if err != nil {
		ctx.Error(err.Error(), http.StatusBadRequest)

		return
	}

	ctx.SetContentType("application/json")
	_ = encoder.Encode(&VerificationResponse{
		Me:       token.Me,
		ClientID: token.ClientID,
		Scope:    token.Scope,
	})
}

func (h *Handler) Update(ctx *http.RequestCtx) {
	if ctx.PostArgs().Has("action") {
		h.Revoke(ctx)

		return
	}

	h.Exchange(ctx)
}

func (r *ExchangeRequest) bind(ctx *http.RequestCtx) error {
	if r.GrantType = string(ctx.PostArgs().Peek("grant_type")); r.GrantType != "authorization_code" {
		return model.Error{
			Code:        model.ErrInvalidRequest.Code,
			Description: "'grant_type' must be 'authorization_code'",
		}
	}

	if r.Code = string(ctx.PostArgs().Peek("code")); r.Code == "" {
		return model.Error{
			Code:        model.ErrInvalidRequest.Code,
			Description: "'code' query is required",
		}
	}

	if r.ClientID = string(ctx.PostArgs().Peek("client_id")); r.ClientID == "" {
		return model.Error{
			Code:        model.ErrInvalidRequest.Code,
			Description: "'client_id' query is required",
		}
	}

	if r.RedirectURI = string(ctx.PostArgs().Peek("redirect_uri")); r.RedirectURI == "" {
		return model.Error{
			Code:        model.ErrInvalidRequest.Code,
			Description: "'redirect_uri' query is required",
		}
	}

	r.CodeVerifier = string(ctx.PostArgs().Peek("code_verifier"))

	return nil
}

func (h *Handler) Exchange(ctx *http.RequestCtx) {
	encoder := json.NewEncoder(ctx)
	req := new(ExchangeRequest)

	if err := req.bind(ctx); err != nil {
		ctx.Error(err.Error(), http.StatusBadRequest)

		return
	}

	token, err := h.useCase.Exchange(ctx, &model.ExchangeRequest{
		ClientID:     req.ClientID,
		Code:         req.Code,
		CodeVerifier: req.CodeVerifier,
		RedirectURI:  req.RedirectURI,
	})
	if err != nil {
		ctx.Error(err.Error(), http.StatusInternalServerError)

		return
	}

	if token == nil {
		ctx.Error(model.ErrUnauthorizedClient.Error(), http.StatusUnauthorized)

		return
	}

	ctx.SetContentType("application/json")
	_ = encoder.Encode(&ExchangeResponse{
		AccessToken: token.AccessToken,
		Me:          token.Me,
		Scope:       token.Scope,
		TokenType:   token.TokenType,
	})
}

func (r *RevokeRequest) bind(ctx *http.RequestCtx) error {
	if r.Action = string(ctx.PostArgs().Peek("action")); r.Action != "revoke" {
		return model.Error{
			Code:        model.ErrInvalidRequest.Code,
			Description: "'action' must be 'revoke'",
		}
	}

	if r.Token = string(ctx.PostArgs().Peek("token")); r.Token == "" {
		return model.Error{
			Code:        model.ErrInvalidRequest.Code,
			Description: "'token' query is required",
		}
	}

	return nil
}

func (h *Handler) Revoke(ctx *http.RequestCtx) {
	req := new(RevokeRequest)
	if err := req.bind(ctx); err != nil {
		ctx.Error(err.Error(), http.StatusBadRequest)

		return
	}

	_ = h.useCase.Revoke(ctx, string(ctx.PostArgs().Peek("token")))

	ctx.SuccessString("application/json", "{}")
}
