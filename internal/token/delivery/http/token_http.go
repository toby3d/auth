package http

import (
	"bytes"
	"strings"

	"github.com/fasthttp/router"
	json "github.com/goccy/go-json"
	http "github.com/valyala/fasthttp"
	"golang.org/x/xerrors"

	"source.toby3d.me/website/oauth/internal/common"
	"source.toby3d.me/website/oauth/internal/domain"
	"source.toby3d.me/website/oauth/internal/token"
)

type (
	RequestHandler struct {
		useCase token.UseCase
	}

	RevocationRequest struct {
		Action string
		Token  string
	}

	//nolint: tagliatelle
	VerificationResponse struct {
		Me       string `json:"me"`
		ClientID string `json:"client_id"`
		Scope    string `json:"scope"`
	}

	RevocationResponse struct{}
)

const (
	Action       string = "action"
	ActionRevoke string = "revoke"
)

func NewRequestHandler(useCase token.UseCase) *RequestHandler {
	return &RequestHandler{
		useCase: useCase,
	}
}

func (h *RequestHandler) Register(r *router.Router) {
	r.GET("/token", h.Read)
	r.POST("/token", h.Update)
}

func (h *RequestHandler) Read(ctx *http.RequestCtx) {
	ctx.SetContentType(common.MIMEApplicationJSON)
	ctx.SetStatusCode(http.StatusOK)

	encoder := json.NewEncoder(ctx)
	rawToken := ctx.Request.Header.Peek(http.HeaderAuthorization)

	token, err := h.useCase.Verify(ctx, string(bytes.TrimSpace(bytes.TrimPrefix(rawToken, []byte("Bearer")))))
	if err != nil {
		ctx.SetStatusCode(http.StatusBadRequest)

		if err = encoder.Encode(err); err != nil {
			ctx.Error(err.Error(), http.StatusInternalServerError)
		}

		return
	}

	if token == nil {
		ctx.SetStatusCode(http.StatusUnauthorized)

		return
	}

	if err := encoder.Encode(&VerificationResponse{
		Me:       token.Me,
		ClientID: token.ClientID,
		Scope:    strings.Join(token.Scopes, " "),
	}); err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)

		if err = encoder.Encode(err); err != nil {
			ctx.Error(err.Error(), http.StatusInternalServerError)
		}
	}
}

func (h *RequestHandler) Update(ctx *http.RequestCtx) {
	if strings.EqualFold(string(ctx.FormValue(Action)), ActionRevoke) {
		h.Revocation(ctx)
	}
}

func (h *RequestHandler) Revocation(ctx *http.RequestCtx) {
	ctx.SetContentType(common.MIMEApplicationJSON)
	ctx.SetStatusCode(http.StatusOK)

	encoder := json.NewEncoder(ctx)

	req := new(RevocationRequest)
	if err := req.bind(ctx); err != nil {
		ctx.SetStatusCode(http.StatusBadRequest)

		if err = encoder.Encode(err); err != nil {
			ctx.Error(err.Error(), http.StatusInternalServerError)
		}

		return
	}

	if err := h.useCase.Revoke(ctx, req.Token); err != nil {
		ctx.SetStatusCode(http.StatusBadRequest)

		if err = encoder.Encode(err); err != nil {
			ctx.Error(err.Error(), http.StatusInternalServerError)
		}

		return
	}

	if err := encoder.Encode(&RevocationResponse{}); err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)

		if err = encoder.Encode(err); err != nil {
			ctx.Error(err.Error(), http.StatusInternalServerError)
		}
	}
}

func (r *RevocationRequest) bind(ctx *http.RequestCtx) error {
	if r.Action = string(ctx.FormValue(Action)); !strings.EqualFold(r.Action, ActionRevoke) {
		return domain.Error{
			Code:        "invalid_request",
			Description: "request MUST contain 'action' key with value 'revoke'",
			URI:         "https://indieauth.spec.indieweb.org/#token-revocation-request",
			Frame:       xerrors.Caller(1),
		}
	}

	if r.Token = string(ctx.FormValue("token")); r.Token == "" {
		return domain.Error{
			Code:        "invalid_request",
			Description: "request MUST contain the 'token' key with the valid access token as its value",
			URI:         "https://indieauth.spec.indieweb.org/#token-revocation-request",
			Frame:       xerrors.Caller(1),
		}
	}

	return nil
}
