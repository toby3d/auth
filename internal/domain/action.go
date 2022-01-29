package domain

import (
	"fmt"
	"strconv"
	"strings"
)

// Action represent action for token endpoint supported by IndieAuth.
//
// NOTE(toby3d): Encapsulate enums in structs for extra compile-time safety:
// https://threedots.tech/post/safer-enums-in-go/#struct-based-enums
type Action struct {
	uid string
}

//nolint: gochecknoglobals // NOTE(toby3d): structs cannot be constants
var (
	ActionUndefined = Action{uid: ""}

	// ActionRevoke represent action for revoke token.
	ActionRevoke = Action{uid: "revoke"}

	// ActionTicket represent action for TicketAuth extension.
	ActionTicket = Action{uid: "ticket"}
)

var ErrActionUnknown error = NewError(ErrorCodeInvalidRequest, "unknown action method")

// ParseAction parse string identifier of action into struct enum.
func ParseAction(uid string) (Action, error) {
	switch strings.ToLower(uid) {
	case ActionRevoke.uid:
		return ActionRevoke, nil
	case ActionTicket.uid:
		return ActionTicket, nil
	}

	return ActionUndefined, fmt.Errorf("%w: %s", ErrActionUnknown, uid)
}

// UnmarshalForm implements custom unmarshler for form values.
func (a *Action) UnmarshalForm(v []byte) error {
	action, err := ParseAction(string(v))
	if err != nil {
		return fmt.Errorf("UnmarshalForm: %w", err)
	}

	*a = action

	return nil
}

// UnmarshalJSON implements custom unmarshler for JSON.
func (a *Action) UnmarshalJSON(v []byte) error {
	src, err := strconv.Unquote(string(v))
	if err != nil {
		return fmt.Errorf("UnmarshalJSON: %w", err)
	}

	action, err := ParseAction(src)
	if err != nil {
		return fmt.Errorf("UnmarshalJSON: %w", err)
	}

	*a = action

	return nil
}

// String returns string representation of action.
func (a Action) String() string {
	return a.uid
}
