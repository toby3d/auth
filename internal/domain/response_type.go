package domain

import (
	"fmt"
	"strconv"
	"strings"

	"source.toby3d.me/toby3d/auth/internal/common"
)

// NOTE(toby3d): Encapsulate enums in structs for extra compile-time safety:
// https://threedots.tech/post/safer-enums-in-go/#struct-based-enums
type ResponseType struct {
	uid string
}

//nolint: gochecknoglobals // structs cannot be constants
var (
	ResponseTypeUnd = ResponseType{uid: ""} // "und"

	// Deprecated(toby3d): Only accept response_type=code requests, and for
	// backwards-compatible support, treat response_type=id requests as
	// response_type=code requests:
	// https://aaronparecki.com/2020/12/03/1/indieauth-2020#response-type
	ResponseTypeID = ResponseType{uid: "id"} // "id"

	// Indicates to the authorization server that an authorization code
	// should be returned as the response:
	// https://indieauth.net/source/#authorization-request-li-1
	ResponseTypeCode = ResponseType{uid: "code"} // "code"
)

var ErrResponseTypeUnknown error = NewError(
	ErrorCodeUnsupportedResponseType,
	"unknown grant type",
	"https://indieauth.net/source/#authorization-request",
)

// ParseResponseType parse string as response type struct enum.
func ParseResponseType(uid string) (ResponseType, error) {
	switch strings.ToLower(uid) {
	case ResponseTypeCode.uid:
		return ResponseTypeCode, nil
	case ResponseTypeID.uid:
		return ResponseTypeID, nil
	}

	return ResponseTypeUnd, fmt.Errorf("%w: %s", ErrResponseTypeUnknown, uid)
}

// UnmarshalForm implements custom unmarshler for form values.
func (rt *ResponseType) UnmarshalForm(src []byte) error {
	responseType, err := ParseResponseType(string(src))
	if err != nil {
		return fmt.Errorf("ResponseType: UnmarshalForm: %w", err)
	}

	*rt = responseType

	return nil
}

// UnmarshalJSON implements custom unmarshler for JSON.
func (rt *ResponseType) UnmarshalJSON(v []byte) error {
	uid, err := strconv.Unquote(string(v))
	if err != nil {
		return fmt.Errorf("ResponseType: UnmarshalJSON: %w", err)
	}

	responseType, err := ParseResponseType(uid)
	if err != nil {
		return fmt.Errorf("ResponseType: UnmarshalJSON: %w", err)
	}

	*rt = responseType

	return nil
}

func (rt ResponseType) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Quote(rt.uid)), nil
}

// String returns string representation of response type.
func (rt ResponseType) String() string {
	if rt.uid != "" {
		return rt.uid
	}

	return common.Und
}

func (rt ResponseType) GoString() string {
	return "domain.ResponseType(" + rt.String() + ")"
}
