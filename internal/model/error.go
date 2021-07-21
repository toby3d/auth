package model

import (
	"fmt"

	"golang.org/x/xerrors"
)

// TODO(toby3d): make more informative errors.
// See https://indieauth.spec.indieweb.org/#authorization-request
type Error struct {
	Code        string        `json:"error"`
	Description string        `json:"error_description,omitempty"`
	URI         string        `json:"error_uri,omitempty"`
	Frame       xerrors.Frame `json:"-"`
}

var (
	ErrInvalidRequest Error = Error{
		Code:        "invalid_request",
		Description: "the request is missing a required parameter, includes an invalid parameter value, or is otherwise malformed",
	}
	ErrUnauthorizedClient Error = Error{
		Code:        "unauthorized_client",
		Description: "the client is not authorized to request an authorization code using this method",
	}
	ErrAccessDenied Error = Error{
		Code:        "access_denied",
		Description: "",
	}
	ErrUnsupportedResponseType Error = Error{
		Code:        "unsupported_response_type",
		Description: "the authorization server does not support obtaining an authorization code using this method",
	}
	ErrInvalidScope Error = Error{
		Code:        "invalid_scope",
		Description: "the requested scope is invalid, unknown, or malformed",
	}
	ErrServerError Error = Error{
		Code:        "server_error",
		Description: "the authorization server encountered an unexpected condition which prevented it from fulfilling the request",
	}
	ErrTemporarilyUnavailable Error = Error{
		Code:        "temporarily_unavailable",
		Description: "the authorization server is currently unable to handle the request due to a temporary overloading or maintenance of the server",
	}
)

func (e Error) FormatError(p xerrors.Printer) error {
	p.Printf("%s: %s", e.Code, e.Description)
	e.Frame.Format(p)

	return nil
}

func (e Error) Format(s fmt.State, r rune) {
	xerrors.FormatError(e, s, r)
}

func (e Error) Error() string {
	return fmt.Sprint(e)
}
