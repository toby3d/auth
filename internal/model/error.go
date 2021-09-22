package model

import (
	"fmt"

	"golang.org/x/xerrors"
)

type Error struct {
	Code        string        `json:"error"`
	Description string        `json:"error_description,omitempty"`
	URI         string        `json:"error_uri,omitempty"`
	Frame       xerrors.Frame `json:"-"`
}

const (
	ErrAccessDenied            string = "access_denied"
	ErrInvalidClient           string = "invalid_client"
	ErrInvalidGrant            string = "invalid_grant"
	ErrInvalidRequest          string = "invalid_request"
	ErrInvalidScope            string = "invalid_scope"
	ErrInvalidToken            string = "invalid_token"
	ErrServerError             string = "server_error"
	ErrTemporarilyUnavailable  string = "temporarily_unavailable"
	ErrUnauthorizedClient      string = "unauthorized_client"
	ErrUnsupportedResponseType string = "unsupported_response_type"
)

const errorColor string = "\033[31m"

func (e Error) Error() string {
	return fmt.Sprint(e)
}

func (e Error) Format(s fmt.State, r rune) {
	xerrors.FormatError(e, s, r)
}

func (e Error) FormatError(p xerrors.Printer) error {
	p.Print(errorColor, e.Code)

	if e.Description != "" {
		p.Printf(": %s", e.Description)
	}

	if e.URI != "" {
		p.Printf("%4s", e.URI)
	}

	if p.Detail() {
		e.Frame.Format(p)
	}

	return nil
}
