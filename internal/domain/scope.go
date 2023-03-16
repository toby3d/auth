package domain

import (
	"fmt"
	"strconv"
	"strings"

	"source.toby3d.me/toby3d/auth/internal/common"
)

// Scope represent single token scope supported by IndieAuth.
//
// NOTE(toby3d): Encapsulate enums in structs for extra compile-time safety:
// https://threedots.tech/post/safer-enums-in-go/#struct-based-enums
type Scope struct {
	scope string
}

var ErrScopeUnknown error = NewError(ErrorCodeInvalidRequest, "unknown scope", "https://indieweb.org/scope")

//nolint:gochecknoglobals // structs cannot be constants
var (
	ScopeUnd = Scope{scope: ""} // "und"

	// https://indieweb.org/scope#Micropub_Scopes
	ScopeCreate   = Scope{scope: "create"}   // "create"
	ScopeDelete   = Scope{scope: "delete"}   // "delete"
	ScopeDraft    = Scope{scope: "draft"}    // "draft"
	ScopeMedia    = Scope{scope: "media"}    // "media"
	ScopeUndelete = Scope{scope: "undelete"} // "undelete"
	ScopeUpdate   = Scope{scope: "update"}   // "update"

	// https://indieweb.org/scope#Microsub_Scopes
	ScopeBlock    = Scope{scope: "block"}    // "block"
	ScopeChannels = Scope{scope: "channels"} // "channels"
	ScopeFollow   = Scope{scope: "follow"}   // "follow"
	ScopeMute     = Scope{scope: "mute"}     // "mute"
	ScopeRead     = Scope{scope: "read"}     // "read"

	// This scope requests access to the user's default profile information
	// which include the following properties: name, photo, url.
	//
	// NOTE(toby3d): https://indieauth.net/source/#profile-information
	ScopeProfile = Scope{scope: "profile"} // "profile"

	// This scope requests access to the user's email address in the
	// following property: email.
	//
	// Note that because the profile scope is required when requesting
	// profile information, the email scope cannot be requested on its own
	// and must be requested along with the profile scope if desired.
	//
	// NOTE(toby3d): https://indieauth.net/source/#profile-information
	ScopeEmail = Scope{scope: "email"} // "email"
)

//nolint:gochecknoglobals // maps cannot be constants
var uidsScopes = map[string]Scope{
	ScopeBlock.scope:    ScopeBlock,
	ScopeChannels.scope: ScopeChannels,
	ScopeCreate.scope:   ScopeCreate,
	ScopeDelete.scope:   ScopeDelete,
	ScopeDraft.scope:    ScopeDraft,
	ScopeEmail.scope:    ScopeEmail,
	ScopeFollow.scope:   ScopeFollow,
	ScopeMedia.scope:    ScopeMedia,
	ScopeMute.scope:     ScopeMute,
	ScopeProfile.scope:  ScopeProfile,
	ScopeRead.scope:     ScopeRead,
	ScopeUndelete.scope: ScopeUndelete,
	ScopeUpdate.scope:   ScopeUpdate,
}

// ParseScope parses scope slug into Scope domain.
func ParseScope(uid string) (Scope, error) {
	if scope, ok := uidsScopes[strings.ToLower(uid)]; ok {
		return scope, nil
	}

	return ScopeUnd, fmt.Errorf("%w: %s", ErrScopeUnknown, uid)
}

func (s *Scope) UnmarshalJSON(v []byte) error {
	src, err := strconv.Unquote(string(v))
	if err != nil {
		return fmt.Errorf("Scope: UnmarshalJSON: cannot unquote string: %w", err)
	}

	out, err := ParseScope(src)
	if err != nil {
		return fmt.Errorf("Scopes: UnmarshalJSON: cannot parse scope: %w", err)
	}

	*s = out

	return nil
}

func (s Scope) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Quote(s.scope)), nil
}

// String returns string representation of scope.
func (s Scope) String() string {
	if s.scope != "" {
		return s.scope
	}

	return common.Und
}

func (s Scope) GoString() string {
	return "domain.Scope(" + s.String() + ")"
}
