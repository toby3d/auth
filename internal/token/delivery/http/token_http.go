package http

import (
	"strings"

	"github.com/fasthttp/router"
	"github.com/goccy/go-json"
	http "github.com/valyala/fasthttp"

	"source.toby3d.me/website/oauth/internal/common"
	"source.toby3d.me/website/oauth/internal/model"
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
	r.POST("/token", h.Update)
}

func (h *RequestHandler) Update(ctx *http.RequestCtx) {
	ctx.SetStatusCode(http.StatusBadRequest)

	switch string(ctx.FormValue(Action)) {
	case ActionRevoke:
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
		encoder.Encode(err)

		return
	}

	if err := h.useCase.Revoke(ctx, req.Token); err != nil {
		ctx.SetStatusCode(http.StatusBadRequest)
		encoder.Encode(err)

		return
	}

	if err := encoder.Encode(RevocationResponse{}); err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		encoder.Encode(err)

		return
	}
}

func (r *RevocationRequest) bind(ctx *http.RequestCtx) error {
	if r.Action = string(ctx.FormValue(Action)); !strings.EqualFold(r.Action, ActionRevoke) {
		return model.Error{
			Code:        model.ErrInvalidRequest,
			Description: "request MUST contain 'action' key with value 'revoke'",
			URI:         "https://indieauth.spec.indieweb.org/#token-revocation-request",
		}
	}

	if r.Token = string(ctx.FormValue("token")); r.Token == "" {
		return model.Error{
			Code:        model.ErrInvalidRequest,
			Description: "request MUST contain the 'token' key with the valid access token as its value",
			URI:         "https://indieauth.spec.indieweb.org/#token-revocation-request",
		}
	}

	return nil
}
