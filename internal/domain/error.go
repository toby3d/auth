package domain

import (
	"fmt"
	"strconv"

	http "github.com/valyala/fasthttp"
	"golang.org/x/xerrors"
)

type (
	// Error describes the format of a typical IndieAuth error.
	Error struct {
		// A single error code.
		Code ErrorCode `json:"error"`

		// Human-readable ASCII text providing additional information, used to
		// assist the client developer in understanding the error that occurred.
		Description string `json:"error_description,omitempty"` //nolint: tagliatelle

		// A URI identifying a human-readable web page with information about
		// the error, used to provide the client developer with additional
		// information about the error.
		URI string `json:"error_uri,omitempty"` //nolint: tagliatelle

		// REQUIRED if a "state" parameter was present in the client
		// authorization request. The exact value received from the
		// client.
		State string `json:"-"`

		frame xerrors.Frame `json:"-"`
	}

	// ErrorCode represent error code described in RFC 6749.
	ErrorCode struct {
		uid string
	}
)

var (
	ErrorCodeUndefined ErrorCode = ErrorCode{uid: ""}

	// RFC 6749 section 4.1.2.1: The resource owner or authorization server
	// denied the request.
	ErrorCodeAccessDenied ErrorCode = ErrorCode{uid: "access_denied"}

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
	ErrorCodeInvalidClient ErrorCode = ErrorCode{uid: "invalid_client"}

	// RFC 6749 section 5.2: The provided authorization grant (e.g.,
	// authorization code, resource owner credentials) or refresh token is
	// invalid, expired, revoked, does not match the redirection URI used in
	// the authorization request, or was issued to another client.
	ErrorCodeInvalidGrant ErrorCode = ErrorCode{uid: "invalid_grant"}

	// RFC 6749 section 4.1.2.1: The request is missing a required
	// parameter, includes an invalid parameter value, includes a parameter
	// more than once, or is otherwise malformed.
	//
	// RFC 6749 section 5.2: The request is missing a required parameter,
	// includes an unsupported parameter value (other than grant type),
	// repeats a parameter, includes multiple credentials, utilizes more
	// than one mechanism for authenticating the client, or is otherwise
	// malformed.
	ErrorCodeInvalidRequest ErrorCode = ErrorCode{uid: "invalid_request"}

	// RFC 6749 section 4.1.2.1: The requested scope is invalid, unknown, or
	// malformed.
	//
	// RFC 6749 section 5.2: The requested scope is invalid, unknown,
	// malformed, or exceeds the scope granted by the resource owner.
	ErrorCodeInvalidScope ErrorCode = ErrorCode{uid: "invalid_scope"}

	// RFC 6749 section 4.1.2.1: The authorization server encountered an
	// unexpected condition that prevented it from fulfilling the request.
	// (This error code is needed because a 500 Internal Server Error HTTP
	// status code cannot be returned to the client via an HTTP redirect.)
	ErrorCodeServerError ErrorCode = ErrorCode{uid: "server_error"}

	// RFC 6749 section 4.1.2.1: The authorization server is currently
	// unable to handle the request due to a temporary overloading or
	// maintenance of the server. (This error code is needed because a 503
	// Service Unavailable HTTP status code cannot be returned to the client
	// via an HTTP redirect.)
	ErrorCodeTemporarilyUnavailable ErrorCode = ErrorCode{uid: "temporarily_unavailable"}

	// RFC 6749 section 4.1.2.1: The client is not authorized to request an
	// authorization code using this method.
	//
	// RFC 6749 section 5.2: The authenticated client is not authorized to
	// use this authorization grant type.
	ErrorCodeUnauthorizedClient ErrorCode = ErrorCode{uid: "unauthorized_client"}

	// RFC 6749 section 5.2: The authorization grant type is not supported
	// by the authorization server.
	ErrorCodeUnsupportedGrantType ErrorCode = ErrorCode{uid: "unsupported_grant_type"}

	// RFC 6749 section 4.1.2.1: The authorization server does not support
	// obtaining an authorization code using this method.
	ErrorCodeUnsupportedResponseType ErrorCode = ErrorCode{uid: "unsupported_response_type"}
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
func (e Error) Format(s fmt.State, r rune) {
	xerrors.FormatError(e, s, r)
}

// FormatError prints the receiver's error, if any.
func (e Error) FormatError(p xerrors.Printer) error {
	p.Print(e.Code)

	if e.Description != "" {
		p.Print(": ", e.Description)
	}

	if !p.Detail() {
		return nil
	}

	e.frame.Format(p)

	return nil
}

// SetReirectURI sets fasthttp.QueryArgs with the request state, code,
// description and error URI in the provided fasthttp.URI.
func (e Error) SetReirectURI(u *http.URI) {
	if u == nil {
		return
	}

	for k, v := range map[string]string{
		"error":             e.Code.String(),
		"error_description": e.Description,
		"error_uri":         e.URI,
		"state":             e.State,
	} {
		if v == "" {
			continue
		}

		u.QueryArgs().Set(k, v)
	}
}

// NewError creates a new Error with the stack pointing to the function call
// line.
//
// If no code or ErrorCodeUndefined is provided, ErrorCodeAccessDenied will be
// used instead.
func NewError(code ErrorCode, description string) *Error {
	if code == ErrorCodeUndefined {
		code = ErrorCodeAccessDenied
	}

	return &Error{
		Code:        code,
		Description: description,
		URI:         "",
		State:       "",
		frame:       xerrors.Caller(1),
	}
}
