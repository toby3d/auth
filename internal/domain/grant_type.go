//nolint: dupl
package domain

import (
	"fmt"
	"strconv"
	"strings"
)

// GrantType represent fixed grant_type parameter.
//
// NOTE(toby3d): Encapsulate enums in structs for extra compile-time safety:
// https://threedots.tech/post/safer-enums-in-go/#struct-based-enums
type GrantType struct {
	uid string
}

//nolint: gochecknoglobals // structs cannot be constants
var (
	GrantTypeUndefined         = GrantType{uid: ""}
	GrantTypeAuthorizationCode = GrantType{uid: "authorization_code"}

	// TicketAuth extension.
	GrantTypeTicket = GrantType{uid: "ticket"}
)

var ErrGrantTypeUnknown error = NewError(
	ErrorCodeInvalidGrant,
	"unknown grant type",
	"",
)

// ParseGrantType parse grant_type value as GrantType struct enum.
func ParseGrantType(uid string) (GrantType, error) {
	switch strings.ToLower(uid) {
	case GrantTypeAuthorizationCode.uid:
		return GrantTypeAuthorizationCode, nil
	case GrantTypeTicket.uid:
		return GrantTypeTicket, nil
	}

	return GrantTypeUndefined, fmt.Errorf("%w: %s", ErrGrantTypeUnknown, uid)
}

// UnmarshalForm implements custom unmarshler for form values.
func (gt *GrantType) UnmarshalForm(src []byte) error {
	responseType, err := ParseGrantType(string(src))
	if err != nil {
		return fmt.Errorf("UnmarshalForm: %w", err)
	}

	*gt = responseType

	return nil
}

// UnmarshalJSON implements custom unmarshler for JSON.
func (gt *GrantType) UnmarshalJSON(v []byte) error {
	src, err := strconv.Unquote(string(v))
	if err != nil {
		return fmt.Errorf("UnmarshalJSON: %w", err)
	}

	responseType, err := ParseGrantType(src)
	if err != nil {
		return fmt.Errorf("UnmarshalJSON: %w", err)
	}

	*gt = responseType

	return nil
}

// String returns string representation of grant type.
func (gt GrantType) String() string {
	return gt.uid
}
