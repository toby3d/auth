package domain

import (
	"fmt"
	"strconv"
	"strings"

	"source.toby3d.me/toby3d/auth/internal/common"
)

// GrantType represent fixed grant_type parameter.
//
// NOTE(toby3d): Encapsulate enums in structs for extra compile-time safety:
// https://threedots.tech/post/safer-enums-in-go/#struct-based-enums
type GrantType struct {
	uid string
}

//nolint:gochecknoglobals // structs cannot be constants
var (
	GrantTypeUnd               = GrantType{uid: ""}                   // "und"
	GrantTypeAuthorizationCode = GrantType{uid: "authorization_code"} // "authorization_code"
	GrantTypeRefreshToken      = GrantType{uid: "refresh_token"}      // "refresh_token"

	// TicketAuth extension.
	GrantTypeTicket = GrantType{uid: "ticket"}
)

var ErrGrantTypeUnknown error = NewError(
	ErrorCodeInvalidGrant,
	"unknown grant type",
	"",
)

//nolint:gochecknoglobals // maps cannot be constants
var uidsGrantTypes = map[string]GrantType{
	GrantTypeAuthorizationCode.uid: GrantTypeAuthorizationCode,
	GrantTypeRefreshToken.uid:      GrantTypeRefreshToken,
	GrantTypeTicket.uid:            GrantTypeTicket,
}

// ParseGrantType parse grant_type value as GrantType struct enum.
func ParseGrantType(uid string) (GrantType, error) {
	if grantType, ok := uidsGrantTypes[strings.ToLower(uid)]; ok {
		return grantType, nil
	}

	return GrantTypeUnd, fmt.Errorf("%w: %s", ErrGrantTypeUnknown, uid)
}

// UnmarshalForm implements custom unmarshler for form values.
func (gt *GrantType) UnmarshalForm(src []byte) error {
	responseType, err := ParseGrantType(string(src))
	if err != nil {
		return fmt.Errorf("GrantType: UnmarshalForm: %w", err)
	}

	*gt = responseType

	return nil
}

// UnmarshalJSON implements custom unmarshler for JSON.
func (gt *GrantType) UnmarshalJSON(v []byte) error {
	src, err := strconv.Unquote(string(v))
	if err != nil {
		return fmt.Errorf("GrantType: UnmarshalJSON: %w", err)
	}

	responseType, err := ParseGrantType(src)
	if err != nil {
		return fmt.Errorf("GrantType: UnmarshalJSON: %w", err)
	}

	*gt = responseType

	return nil
}

func (gt GrantType) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Quote(gt.uid)), nil
}

// String returns string representation of grant type.
func (gt GrantType) String() string {
	if gt.uid != "" {
		return gt.uid
	}

	return common.Und
}

func (gt GrantType) GoString() string {
	return "domain.GrantType(" + gt.String() + ")"
}
