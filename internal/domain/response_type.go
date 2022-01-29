package domain

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// NOTE(toby3d): Encapsulate enums in structs for extra compile-time safety:
// https://threedots.tech/post/safer-enums-in-go/#struct-based-enums
type ResponseType struct {
	uid string
}

//nolint: gochecknoglobals // NOTE(toby3d): structs cannot be constants
var (
	ResponseTypeUndefined ResponseType = ResponseType{uid: ""}

	// Deprecated(toby3d): Only accept response_type=code requests, and for
	// backwards-compatible support, treat response_type=id requests as
	// response_type=code requests:
	// https://aaronparecki.com/2020/12/03/1/indieauth-2020#response-type
	ResponseTypeID ResponseType = ResponseType{uid: "id"}

	// Indicates to the authorization server that an authorization code
	// should be returned as the response:
	// https://indieauth.net/source/#authorization-request-li-1
	ResponseTypeCode ResponseType = ResponseType{uid: "code"}
)

var ErrResponseTypeUnknown error = errors.New("unknown grant type")

// ParseResponseType parse string as response type struct enum.
func ParseResponseType(uid string) (ResponseType, error) {
	switch strings.ToLower(uid) {
	case ResponseTypeCode.uid:
		return ResponseTypeCode, nil
	case ResponseTypeID.uid:
		return ResponseTypeID, nil
	}

	return ResponseTypeUndefined, fmt.Errorf("%w: %s", ErrResponseTypeUnknown, uid)
}

// UnmarshalForm implements custom unmarshler for form values.
func (rt *ResponseType) UnmarshalForm(src []byte) error {
	responseType, err := ParseResponseType(string(src))
	if err != nil {
		return fmt.Errorf("UnmarshalForm: %w", err)
	}

	*rt = responseType

	return nil
}

// UnmarshalJSON implements custom unmarshler for JSON.
func (rt *ResponseType) UnmarshalJSON(v []byte) error {
	uid, err := strconv.Unquote(string(v))
	if err != nil {
		return fmt.Errorf("UnmarshalJSON: %w", err)
	}

	responseType, err := ParseResponseType(string(uid))
	if err != nil {
		return fmt.Errorf("UnmarshalJSON: %w", err)
	}

	*rt = responseType

	return nil
}

// String returns string representation of response type.
func (rt ResponseType) String() string {
	return rt.uid
}
