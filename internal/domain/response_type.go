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
	slug string
}

//nolint: gochecknoglobals // NOTE(toby3d): structs cannot be constants
var (
	ResponseTypeUndefined ResponseType = ResponseType{
		slug: "",
	}

	// Deprecated(toby3d): Only accept response_type=code requests, and for
	// backwards-compatible support, treat response_type=id requests as
	// response_type=code requests:
	// https://aaronparecki.com/2020/12/03/1/indieauth-2020#response-type
	ResponseTypeID ResponseType = ResponseType{
		slug: "id",
	}

	// Indicates to the authorization server that an authorization code
	// should be returned as the response:
	// https://indieauth.net/source/#authorization-request-li-1
	ResponseTypeCode ResponseType = ResponseType{
		slug: "code",
	}
)

var ErrResponseTypeUnknown = errors.New("unknown grant type")

func ParseResponseType(slug string) (ResponseType, error) {
	switch strings.ToLower(slug) {
	case ResponseTypeCode.slug:
		return ResponseTypeCode, nil
	case ResponseTypeID.slug:
		return ResponseTypeID, nil
	}

	return ResponseTypeUndefined, fmt.Errorf("%w: %s", ErrResponseTypeUnknown, slug)
}

func (rt *ResponseType) UnmarshalForm(src []byte) error {
	responseType, err := ParseResponseType(string(src))
	if err != nil {
		return fmt.Errorf("response_type: %w", err)
	}

	*rt = responseType

	return nil
}

func (rt *ResponseType) UnmarshalJSON(v []byte) error {
	src, err := strconv.Unquote(string(v))
	if err != nil {
		return err
	}

	responseType, err := ParseResponseType(string(src))
	if err != nil {
		return fmt.Errorf("response_type: %w", err)
	}

	*rt = responseType

	return nil
}

func (rt ResponseType) String() string {
	return rt.slug
}
