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
	uid string
}

var ErrScopeUnknown error = NewError(ErrorCodeInvalidRequest, "unknown scope", "https://indieweb.org/scope")

//nolint:gochecknoglobals // structs cannot be constants
var (
	ScopeUnd = Scope{uid: ""} // "und"

	// https://indieweb.org/scope#Micropub_Scopes
	ScopeCreate   = Scope{uid: "create"}   // "create"
	ScopeDelete   = Scope{uid: "delete"}   // "delete"
	ScopeDraft    = Scope{uid: "draft"}    // "draft"
	ScopeMedia    = Scope{uid: "media"}    // "media"
	ScopeUndelete = Scope{uid: "undelete"} // "undelete"
	ScopeUpdate   = Scope{uid: "update"}   // "update"

	// https://indieweb.org/scope#Microsub_Scopes
	ScopeBlock    = Scope{uid: "block"}    // "block"
	ScopeChannels = Scope{uid: "channels"} // "channels"
	ScopeFollow   = Scope{uid: "follow"}   // "follow"
	ScopeMute     = Scope{uid: "mute"}     // "mute"
	ScopeRead     = Scope{uid: "read"}     // "read"

	// This scope requests access to the user's default profile information
	// which include the following properties: name, photo, url.
	//
	// NOTE(toby3d): https://indieauth.net/source/#profile-information
	ScopeProfile = Scope{uid: "profile"} // "profile"

	// This scope requests access to the user's email address in the
	// following property: email.
	//
	// Note that because the profile scope is required when requesting
	// profile information, the email scope cannot be requested on its own
	// and must be requested along with the profile scope if desired.
	//
	// NOTE(toby3d): https://indieauth.net/source/#profile-information
	ScopeEmail = Scope{uid: "email"} // "email"
)

//nolint:gochecknoglobals // maps cannot be constants
var uidsScopes = map[string]Scope{
	ScopeBlock.uid:    ScopeBlock,
	ScopeChannels.uid: ScopeChannels,
	ScopeCreate.uid:   ScopeCreate,
	ScopeDelete.uid:   ScopeDelete,
	ScopeDraft.uid:    ScopeDraft,
	ScopeEmail.uid:    ScopeEmail,
	ScopeFollow.uid:   ScopeFollow,
	ScopeMedia.uid:    ScopeMedia,
	ScopeMute.uid:     ScopeMute,
	ScopeProfile.uid:  ScopeProfile,
	ScopeRead.uid:     ScopeRead,
	ScopeUndelete.uid: ScopeUndelete,
	ScopeUpdate.uid:   ScopeUpdate,
}

// ParseScope parses scope slug into Scope domain.
func ParseScope(uid string) (Scope, error) {
	if scope, ok := uidsScopes[strings.ToLower(uid)]; ok {
		return scope, nil
	}

	return ScopeUnd, fmt.Errorf("%w: %s", ErrScopeUnknown, uid)
}

func (s Scope) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Quote(s.uid)), nil
}

// String returns string representation of scope.
func (s Scope) String() string {
	if s.uid != "" {
		return s.uid
	}

	return common.Und
}

func (s Scope) GoString() string {
	return "domain.Scope(" + s.String() + ")"
}
