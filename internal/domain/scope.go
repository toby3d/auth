package domain

import (
	"fmt"
	"strconv"
	"strings"
)

type (
	// Scope represent single token scope supported by IndieAuth.
	//
	// NOTE(toby3d): Encapsulate enums in structs for extra compile-time safety:
	// https://threedots.tech/post/safer-enums-in-go/#struct-based-enums
	Scope struct {
		uid string
	}

	// Scopes represent set of Scope domains.
	Scopes []Scope
)

var ErrScopeUnknown error = NewError(ErrorCodeInvalidRequest, "unknown scope", "https://indieweb.org/scope")

//nolint: gochecknoglobals // structs cannot be constants
var (
	ScopeUndefined = Scope{uid: ""}

	// https://indieweb.org/scope#Micropub_Scopes
	ScopeCreate = Scope{uid: "create"}
	ScopeDelete = Scope{uid: "delete"}
	ScopeDraft  = Scope{uid: "draft"}
	ScopeMedia  = Scope{uid: "media"}
	ScopeUpdate = Scope{uid: "update"}

	// https://indieweb.org/scope#Microsub_Scopes
	ScopeBlock    = Scope{uid: "block"}
	ScopeChannels = Scope{uid: "channels"}
	ScopeFollow   = Scope{uid: "follow"}
	ScopeMute     = Scope{uid: "mute"}
	ScopeRead     = Scope{uid: "read"}

	// This scope requests access to the user's default profile information
	// which include the following properties: name, photo, url.
	//
	// NOTE(toby3d): https://indieauth.net/source/#profile-information
	ScopeProfile = Scope{uid: "profile"}

	// This scope requests access to the user's email address in the
	// following property: email.
	//
	// Note that because the profile scope is required when requesting
	// profile information, the email scope cannot be requested on its own
	// and must be requested along with the profile scope if desired.
	//
	// NOTE(toby3d): https://indieauth.net/source/#profile-information
	ScopeEmail = Scope{uid: "email"}
)

//nolint: gochecknoglobals // maps cannot be constants
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
	ScopeUpdate.uid:   ScopeUpdate,
}

// ParseScope parses scope slug into Scope domain.
func ParseScope(uid string) (Scope, error) {
	if scope, ok := uidsScopes[strings.ToLower(uid)]; ok {
		return scope, nil
	}

	return ScopeUndefined, fmt.Errorf("%w: %s", ErrScopeUnknown, uid)
}

func (s Scope) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Quote(s.uid)), nil
}

// String returns string representation of scope.
func (s Scope) String() string {
	return s.uid
}

// UnmarshalForm implements custom unmarshler for form values.
func (s *Scopes) UnmarshalForm(v []byte) error {
	scopes := make(Scopes, 0)

	for _, rawScope := range strings.Fields(string(v)) {
		scope, err := ParseScope(rawScope)
		if err != nil {
			return fmt.Errorf("Scopes: UnmarshalForm: %w", err)
		}

		if scopes.Has(scope) {
			continue
		}

		scopes = append(scopes, scope)
	}

	*s = scopes

	return nil
}

// UnmarshalJSON implements custom unmarshler for JSON.
func (s *Scopes) UnmarshalJSON(v []byte) error {
	src, err := strconv.Unquote(string(v))
	if err != nil {
		return fmt.Errorf("Scopes: UnmarshalJSON: %w", err)
	}

	result := make(Scopes, 0)

	for _, rawScope := range strings.Fields(src) {
		scope, err := ParseScope(rawScope)
		if err != nil {
			return fmt.Errorf("Scopes: UnmarshalJSON: %w", err)
		}

		if result.Has(scope) {
			continue
		}

		result = append(result, scope)
	}

	*s = result

	return nil
}

// UnmarshalJSON implements custom marshler for JSON.
func (s Scopes) MarshalJSON() ([]byte, error) {
	scopes := make([]string, len(s))

	for i := range s {
		scopes[i] = s[i].String()
	}

	return []byte(strconv.Quote(strings.Join(scopes, " "))), nil
}

// String returns string representation of scopes.
func (s Scopes) String() string {
	scopes := make([]string, len(s))

	for i := range s {
		scopes[i] = s[i].String()
	}

	return strings.Join(scopes, " ")
}

// IsEmpty returns true if the set does not contain valid scope.
func (s Scopes) IsEmpty() bool {
	for i := range s {
		if s[i] == ScopeUndefined {
			continue
		}

		return false
	}

	return true
}

// Has check what input scope contains in current scopes collection.
func (s Scopes) Has(scope Scope) bool {
	for i := range s {
		if s[i] != scope {
			continue
		}

		return true
	}

	return false
}
