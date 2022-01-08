package domain

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

type (
	// NOTE(toby3d): https://threedots.tech/post/safer-enums-in-go/#struct-based-enums
	Scope struct {
		slug string
	}

	// Scopes represent set of Scope domains.
	Scopes []Scope
)

var ErrScopeUnknown = errors.New("unknown scope")

//nolint: gochecknoglobals // NOTE(toby3d): structs cannot be constants
var (
	ScopeUndefined = Scope{slug: ""}

	// https://indieweb.org/scope#Micropub_Scopes
	ScopeCreate = Scope{slug: "create"}
	ScopeDelete = Scope{slug: "delete"}
	ScopeDraft  = Scope{slug: "draft"}
	ScopeMedia  = Scope{slug: "media"}
	ScopeUpdate = Scope{slug: "update"}

	// https://indieweb.org/scope#Microsub_Scopes
	ScopeBlock    = Scope{slug: "block"}
	ScopeChannels = Scope{slug: "channels"}
	ScopeFollow   = Scope{slug: "follow"}
	ScopeMute     = Scope{slug: "mute"}
	ScopeRead     = Scope{slug: "read"}

	// This scope requests access to the user's default profile information
	// which include the following properties: name, `photo, url.
	//
	// NOTE(toby3d): https://indieauth.net/source/#profile-information
	ScopeProfile = Scope{
		slug: "profile",
	}

	// This scope requests access to the user's email address in the
	// following property: email.
	//
	// Note that because the profile scope is required when requesting
	// profile information, the email scope cannot be requested on its own
	// and must be requested along with the profile scope if desired.
	//
	// NOTE(toby3d): https://indieauth.net/source/#profile-information
	ScopeEmail = Scope{
		slug: "email",
	}
)

//nolint: gochecknoglobals // NOTE(toby3d): maps cannot be constants
var slugsScopes = map[string]Scope{
	ScopeBlock.slug:    ScopeBlock,
	ScopeChannels.slug: ScopeChannels,
	ScopeCreate.slug:   ScopeCreate,
	ScopeDelete.slug:   ScopeDelete,
	ScopeDraft.slug:    ScopeDraft,
	ScopeEmail.slug:    ScopeEmail,
	ScopeFollow.slug:   ScopeFollow,
	ScopeMedia.slug:    ScopeMedia,
	ScopeMute.slug:     ScopeMute,
	ScopeProfile.slug:  ScopeProfile,
	ScopeRead.slug:     ScopeRead,
	ScopeUpdate.slug:   ScopeUpdate,
}

// ParseScope parses scope slug into Scope domain.
func ParseScope(slug string) (Scope, error) {
	if scope, ok := slugsScopes[strings.ToLower(slug)]; ok {
		return scope, nil
	}

	return ScopeUndefined, fmt.Errorf("%w: %s", ErrScopeUnknown, slug)
}

// UnmarshalForm parses the value of the form key into the Scope domain.
func (s *Scope) UnmarshalForm(v []byte) (err error) {
	scope, err := ParseScope(string(v))
	if err != nil {
		return fmt.Errorf("scope: %w", err)
	}

	*s = scope

	return nil
}

func (s *Scope) UnmarshalJSON(v []byte) error {
	src, err := strconv.Unquote(string(v))
	if err != nil {
		return err
	}

	scope, err := ParseScope(src)
	if err != nil {
		return fmt.Errorf("scope: %w", err)
	}

	*s = scope

	return nil
}

func (s *Scopes) UnmarshalJSON(v []byte) error {
	src, err := strconv.Unquote(string(v))
	if err != nil {
		return err
	}

	result := make([]Scope, 0)

	for _, scope := range strings.Fields(src) {
		s, err := ParseScope(scope)
		if err != nil {
			return fmt.Errorf("scope: %w", err)
		}

		result = append(result, s)
	}

	*s = result

	return nil
}

func (s Scopes) MarshalJSON() ([]byte, error) {
	scopes := make([]string, len(s))

	for i := range s {
		scopes[i] = s[i].String()
	}

	sort.Strings(scopes)

	return []byte(strconv.Quote(strings.Join(scopes, " "))), nil
}

// String returns scope slug as string.
func (s Scope) String() string {
	return s.slug
}

func (s Scopes) String() string {
	scopes := make([]string, len(s))

	for i := range s {
		scopes[i] = s[i].String()
	}

	return strings.Join(scopes, " ")
}
