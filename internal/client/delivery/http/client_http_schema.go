package http

import (
	"errors"
	"net/http"

	"source.toby3d.me/toby3d/auth/internal/domain"
	"source.toby3d.me/toby3d/form"
)

type ClientCallbackRequest struct {
	Error            domain.ErrorCode `form:"error,omitempty"`
	Iss              domain.ClientID  `form:"iss"`
	Code             string           `form:"code"`
	ErrorDescription string           `form:"error_description,omitempty"`
	State            string           `form:"state"`
}

func (req *ClientCallbackRequest) bind(r *http.Request) error {
	indieAuthError := new(domain.Error)

	if err := form.Unmarshal([]byte(r.URL.Query().Encode()), req); err != nil {
		if errors.As(err, indieAuthError) {
			return indieAuthError
		}

		return domain.NewError(domain.ErrorCodeInvalidRequest, err.Error(), "")
	}

	return nil
}
