package domain

import (
	"fmt"
	"strconv"

	http "github.com/valyala/fasthttp"
	"golang.org/x/xerrors"
)

type (
	// Error describes the format of a typical IndieAuth error.
	//nolint: tagliatelle // RFC 6749 section 5.2
	Error struct {
		// A single error code.
		Code ErrorCode `json:"error"`

		// Human-readable ASCII text providing additional information, used to
		// assist the client developer in understanding the error that occurred.
		Description string `json:"error_description,omitempty"`

		// A URI identifying a human-readable web page with information about
		// the error, used to provide the client developer with additional
		// information about the error.
		URI string `json:"error_uri,omitempty"`

		// REQUIRED if a "state" parameter was present in the client
		// authorization request. The exact value received from the
		// client.
		State string `json:"-"`

		frame xerrors.Frame `json:"-"`
	}

	// ErrorCode represent error code described in RFC 6749.
	ErrorCode struct {
		uid    string
		status int
	}
)

var (
	// ErrorCodeUndefined describes an unrecognized error code.
	ErrorCodeUndefined = ErrorCode{
		uid:    "",
		status: 0,
	}

	// ErrorCodeAccessDenied describes the access_denied error code.
	//
	// RFC 6749 section 4.1.2.1: The resource owner or authorization server
	// denied the request.
	ErrorCodeAccessDenied = ErrorCode{
		uid:    "access_denied",
		status: 0, // TODO(toby3d)
	}

	// ErrorCodeInvalidClient describes the invalid_client error code.
	//
	// RFC 6749 section 5.2: Client authentication failed (e.g., unknown
	// client, no client authentication included, or unsupported
	// authentication method).
	//
	// The authorization server MAY return an HTTP 401 (Unauthorized) status
	// code to indicate which HTTP authentication schemes are supported.
	//
	// If the client attempted to authenticate via the "Authorization"
	// request header field, the authorization server MUST respond with an
	// HTTP 401 (Unauthorized) status code and include the
	// "WWW-Authenticate" response header field matching the authentication
	// scheme used by the client.
	ErrorCodeInvalidClient = ErrorCode{
		uid:    "invalid_client",
		status: 0, // TODO(toby3d)
	}

	// ErrorCodeInvalidGrant describes the invalid_grant error code.
	//
	// RFC 6749 section 5.2: The provided authorization grant (e.g.,
	// authorization code, resource owner credentials) or refresh token is
	// invalid, expired, revoked, does not match the redirection URI used in
	// the authorization request, or was issued to another client.
	ErrorCodeInvalidGrant = ErrorCode{
		uid:    "invalid_grant",
		status: 0, // TODO(toby3d)
	}

	// ErrorCodeInvalidRequest describes the invalid_request error code.
	//
	// IndieAuth: The request is not valid.
	//
	// RFC 6749 section 4.1.2.1: The request is missing a required
	// parameter, includes an invalid parameter value, includes a parameter
	// more than once, or is otherwise malformed.
	//
	// RFC 6749 section 5.2: The request is missing a required parameter,
	// includes an unsupported parameter value (other than grant type),
	// repeats a parameter, includes multiple credentials, utilizes more
	// than one mechanism for authenticating the client, or is otherwise
	// malformed.
	ErrorCodeInvalidRequest = ErrorCode{
		uid:    "invalid_request",
		status: http.StatusBadRequest,
	}

	// ErrorCodeInvalidScope describes the invalid_scope error code.
	//
	// RFC 6749 section 4.1.2.1: The requested scope is invalid, unknown, or
	// malformed.
	//
	// RFC 6749 section 5.2: The requested scope is invalid, unknown,
	// malformed, or exceeds the scope granted by the resource owner.
	ErrorCodeInvalidScope = ErrorCode{
		uid:    "invalid_scope",
		status: 0, // TODO(toby3d)
	}

	// ErrorCodeServerError describes the server_error error code.
	//
	// RFC 6749 section 4.1.2.1: The authorization server encountered an
	// unexpected condition that prevented it from fulfilling the request.
	// (This error code is needed because a 500 Internal Server Error HTTP
	// status code cannot be returned to the client via an HTTP redirect.)
	ErrorCodeServerError = ErrorCode{
		uid:    "server_error",
		status: 0, // TODO(toby3d)
	}

	// ErrorCodeTemporarilyUnavailable describes the temporarily_unavailable error code.
	//
	// RFC 6749 section 4.1.2.1: The authorization server is currently
	// unable to handle the request due to a temporary overloading or
	// maintenance of the server. (This error code is needed because a 503
	// Service Unavailable HTTP status code cannot be returned to the client
	// via an HTTP redirect.)
	ErrorCodeTemporarilyUnavailable = ErrorCode{
		uid:    "temporarily_unavailable",
		status: 0, // TODO(toby3d)
	}

	// ErrorCodeUnauthorizedClient describes the unauthorized_client error code.
	//
	// RFC 6749 section 4.1.2.1: The client is not authorized to request an
	// authorization code using this method.
	//
	// RFC 6749 section 5.2: The authenticated client is not authorized to
	// use this authorization grant type.
	ErrorCodeUnauthorizedClient = ErrorCode{
		uid:    "unauthorized_client",
		status: 0, // TODO(toby3d)
	}

	// ErrorCodeUnsupportedGrantType describes the unsupported_grant_type error code.
	//
	// RFC 6749 section 5.2: The authorization grant type is not supported
	// by the authorization server.
	ErrorCodeUnsupportedGrantType = ErrorCode{
		uid:    "unsupported_grant_type",
		status: 0, // TODO(toby3d)
	}

	// ErrorCodeUnsupportedResponseType describes the unsupported_response_type error code.
	//
	// RFC 6749 section 4.1.2.1: The authorization server does not support
	// obtaining an authorization code using this method.
	ErrorCodeUnsupportedResponseType = ErrorCode{
		uid:    "unsupported_response_type",
		status: 0, // TODO(toby3d)
	}

	// ErrorCodeInvalidToken describes the invalid_token error code.
	//
	// IndieAuth: The access token provided is expired, revoked, or invalid.
	ErrorCodeInvalidToken = ErrorCode{
		uid:    "invalid_token",
		status: http.StatusUnauthorized,
	}

	// ErrorCodeInsufficientScope describes the insufficient_scope error code.
	//
	// IndieAuth: The request requires higher privileges than provided.
	ErrorCodeInsufficientScope = ErrorCode{
		uid:    "insufficient_scope",
		status: http.StatusForbidden,
	}
)

// String returns a string representation of the error code.
func (ec ErrorCode) String() string {
	return ec.uid
}

// MarshalJSON encodes the error code into its string representation in JSON.
func (ec ErrorCode) MarshalJSON() ([]byte, error) {
	return []byte(strconv.QuoteToASCII(ec.uid)), nil
}

// Error returns a string representation of the error, satisfying the error
// interface.
func (e Error) Error() string {
	return fmt.Sprint(e)
}

// Format prints the stack as error detail.
func (e Error) Format(state fmt.State, r rune) {
	xerrors.FormatError(e, state, r)
}

// FormatError prints the receiver's error, if any.
func (e Error) FormatError(printer xerrors.Printer) error {
	printer.Print(e.Code)

	if e.Description != "" {
		printer.Print(": ", e.Description)
	}

	if !printer.Detail() {
		return nil
	}

	e.frame.Format(printer)

	return nil
}

// SetReirectURI sets fasthttp.QueryArgs with the request state, code,
// description and error URI in the provided fasthttp.URI.
func (e Error) SetReirectURI(uri *http.URI) {
	if uri == nil {
		return
	}

	for key, val := range map[string]string{
		"error":             e.Code.String(),
		"error_description": e.Description,
		"error_uri":         e.URI,
		"state":             e.State,
	} {
		if val == "" {
			continue
		}

		uri.QueryArgs().Set(key, val)
	}
}

// NewError creates a new Error with the stack pointing to the function call
// line number.
//
// If no code or ErrorCodeUndefined is provided, ErrorCodeAccessDenied will be
// used instead.
func NewError(code ErrorCode, description, uri string, requestState ...string) *Error {
	if code == ErrorCodeUndefined {
		code = ErrorCodeAccessDenied
	}

	var state string
	if len(requestState) > 0 {
		state = requestState[0]
	}

	return &Error{
		Code:        code,
		Description: description,
		URI:         uri,
		State:       state,
		frame:       xerrors.Caller(1),
	}
}
