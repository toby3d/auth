package http

import (
	"errors"
	"net/http"

	"source.toby3d.me/toby3d/auth/internal/domain"
	"source.toby3d.me/toby3d/form"
)

type (
	TicketGenerateRequest struct {
		// The access token should be used when acting on behalf of this URL.
		Subject domain.Me `form:"subject"`

		// The access token will work at this URL.
		Resource domain.URL `form:"resource"`
	}

	TicketExchangeRequest struct {
		// The access token should be used when acting on behalf of this URL.
		Subject domain.Me `form:"subject"`

		// The access token will work at this URL.
		Resource domain.URL `form:"resource"`

		// A random string that can be redeemed for an access token.
		Ticket string `form:"ticket"`
	}
)

func (r *TicketGenerateRequest) bind(req *http.Request) error {
	indieAuthError := new(domain.Error)

	if err := req.ParseForm(); err != nil {
		return domain.NewError(
			domain.ErrorCodeInvalidRequest,
			err.Error(),
			"https://indieauth.net/source/#authorization-request",
		)
	}

	if err := form.Unmarshal([]byte(req.PostForm.Encode()), r); err != nil {
		if errors.As(err, indieAuthError) {
			return indieAuthError
		}

		return domain.NewError(
			domain.ErrorCodeInvalidRequest,
			err.Error(),
			"https://indieweb.org/IndieAuth_Ticket_Auth#Create_the_IndieAuth_ticket",
		)
	}

	return nil
}

func (r *TicketExchangeRequest) bind(req *http.Request) error {
	indieAuthError := new(domain.Error)

	if err := req.ParseForm(); err != nil {
		return domain.NewError(
			domain.ErrorCodeInvalidRequest,
			err.Error(),
			"https://indieauth.net/source/#authorization-request",
		)
	}

	if err := form.Unmarshal([]byte(req.PostForm.Encode()), r); err != nil {
		if errors.As(err, indieAuthError) {
			return indieAuthError
		}

		return domain.NewError(
			domain.ErrorCodeInvalidRequest,
			err.Error(),
			"https://indieweb.org/IndieAuth_Ticket_Auth#Create_the_IndieAuth_ticket",
		)
	}

	if r.Ticket == "" {
		return domain.NewError(
			domain.ErrorCodeInvalidRequest,
			"ticket parameter is required",
			"https://indieweb.org/IndieAuth_Ticket_Auth#Create_the_IndieAuth_ticket",
		)
	}

	return nil
}
