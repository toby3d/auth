package http

import (
	"encoding/json"
	"fmt"

	"github.com/fasthttp/router"
	http "github.com/valyala/fasthttp"
	"golang.org/x/xerrors"

	"source.toby3d.me/toby3d/form"
	"source.toby3d.me/website/oauth/internal/common"
	"source.toby3d.me/website/oauth/internal/domain"
	"source.toby3d.me/website/oauth/internal/ticket"
	"source.toby3d.me/website/oauth/internal/user"
)

type (
	Request struct {
		// A random string that can be redeemed for an access token.
		Ticket string `form:"ticket"`

		// The access token will work at this URL.
		Resource *domain.URL `form:"resource"`

		// The access token should be used when acting on behalf of this URL.
		Subject *domain.Me `form:"subject"`
	}

	RequestHandler struct {
		users   user.UseCase
		tickets ticket.UseCase
	}
)

func NewRequestHandler(tickets ticket.UseCase, users user.UseCase) *RequestHandler {
	return &RequestHandler{
		tickets: tickets,
		users:   users,
	}
}

func (h *RequestHandler) Register(r *router.Router) {
	r.POST("/ticket", h.update)
}

func (h *RequestHandler) update(ctx *http.RequestCtx) {
	ctx.SetContentType(common.MIMEApplicationJSONCharsetUTF8)
	ctx.SetStatusCode(http.StatusOK)

	encoder := json.NewEncoder(ctx)

	req := new(Request)
	if err := req.bind(ctx); err != nil {
		ctx.SetStatusCode(http.StatusBadRequest)
		encoder.Encode(err)

		return
	}

	// TODO(toby3d): fetch token endpoint on Resource URL instead
	u, err := h.users.Fetch(ctx, req.Subject)
	if err != nil {
		ctx.SetStatusCode(http.StatusBadRequest)
		encoder.Encode(domain.Error{
			Code:        "invalid_request",
			Description: err.Error(),
			Frame:       xerrors.Caller(1),
		})

		return
	}

	token, err := h.tickets.Redeem(ctx, u.TokenEndpoint, req.Ticket)
	if err != nil {
		ctx.SetStatusCode(http.StatusBadRequest)
		encoder.Encode(domain.Error{
			Code:        "invalid_request",
			Description: err.Error(),
			Frame:       xerrors.Caller(1),
		})

		return
	}

	// TODO(toby3d): print the result as part of the debugging. Instead, we
	// need to send or save the token to the recipient for later use.
	ctx.SetBodyString(fmt.Sprintf(`{
		"access_token": "%s",
		"token_type": "Bearer",
		"scope": "%s",
		"me": "%s"
	}`, token.AccessToken, token.Scope.String(), token.Me.String()))
}

func (req *Request) bind(ctx *http.RequestCtx) (err error) {
	if err = form.Unmarshal(ctx.Request.PostArgs(), req); err != nil {
		return domain.Error{
			Code:        "invalid_request",
			Description: err.Error(),
			URI:         "https://indieweb.org/IndieAuth_Ticket_Auth#Create_the_IndieAuth_ticket",
			Frame:       xerrors.Caller(1),
		}
	}

	if req.Ticket == "" {
		return domain.Error{
			Code:        "invalid_request",
			Description: "ticket parameter is required",
			URI:         "https://indieweb.org/IndieAuth_Ticket_Auth#Create_the_IndieAuth_ticket",
			Frame:       xerrors.Caller(1),
		}
	}

	if req.Resource == nil {
		return domain.Error{
			Code:        "invalid_request",
			Description: "invalid resource value",
			URI:         "https://indieweb.org/IndieAuth_Ticket_Auth#Create_the_IndieAuth_ticket",
			Frame:       xerrors.Caller(1),
		}
	}

	if req.Subject == nil {
		return domain.Error{
			Code:        "invalid_request",
			Description: "invalid subject value",
			URI:         "https://indieweb.org/IndieAuth_Ticket_Auth#Create_the_IndieAuth_ticket",
			Frame:       xerrors.Caller(1),
		}
	}

	return nil
}