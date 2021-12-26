package domain

import (
	"errors"
	"fmt"
	"strings"
)

// NOTE(toby3d): Encapsulate enums in structs for extra compile-time safety:
// https://threedots.tech/post/safer-enums-in-go/#struct-based-enums
type GrantType struct {
	slug string
}

//nolint: gochecknoglobals // NOTE(toby3d): structs cannot be constants
var (
	GrantTypeUndefined         = GrantType{slug: ""}
	GrantTypeAuthorizationCode = GrantType{slug: "authorization_code"}
)

var ErrGrantTypeUnknown = errors.New("unknown grant type")

func ParseGrantType(slug string) (GrantType, error) {
	if strings.ToLower(slug) == GrantTypeAuthorizationCode.slug {
		return GrantTypeAuthorizationCode, nil
	}

	return GrantTypeUndefined, fmt.Errorf("%w: %s", ErrGrantTypeUnknown, slug)
}

func (gt *GrantType) UnmarshalForm(src []byte) error {
	responseType, err := ParseGrantType(string(src))
	if err != nil {
		return fmt.Errorf("grant_type: %w", err)
	}

	*gt = responseType

	return nil
}

func (gt GrantType) String() string {
	return gt.slug
}
