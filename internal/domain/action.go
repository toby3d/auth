package domain

import (
	"fmt"
	"strconv"

	"source.toby3d.me/toby3d/auth/internal/common"
)

// Action represent action for token endpoint supported by IndieAuth.
//
// NOTE(toby3d): Encapsulate enums in structs for extra compile-time safety:
// https://threedots.tech/post/safer-enums-in-go/#struct-based-enums
type Action struct {
	uid string
}

//nolint:gochecknoglobals // structs cannot be constants
var (
	ActionUnd = Action{uid: ""} // "und"

	// ActionRevoke represent action for revoke token.
	ActionRevoke = Action{uid: "revoke"} // "revoke"

	// ActionTicket represent action for TicketAuth extension.
	ActionTicket = Action{uid: "ticket"} // "ticket"
)

var ErrActionSyntax error = NewError(ErrorCodeInvalidRequest, "unknown action method", "")

//nolint:gochecknoglobals
var uidsActions = map[string]Action{
	ActionRevoke.uid: ActionRevoke,
	ActionTicket.uid: ActionTicket,
}

// ParseAction parse string identifier of action into struct enum.
func ParseAction(uid string) (Action, error) {
	if action, ok := uidsActions[uid]; ok {
		return action, nil
	}

	return ActionUnd, fmt.Errorf("%w: %s", ErrActionSyntax, uid)
}

// UnmarshalForm implements custom unmarshler for form values.
func (a *Action) UnmarshalForm(v []byte) error {
	action, err := ParseAction(string(v))
	if err != nil {
		return fmt.Errorf("Action: UnmarshalForm: %w", err)
	}

	*a = action

	return nil
}

// UnmarshalJSON implements custom unmarshler for JSON.
func (a *Action) UnmarshalJSON(v []byte) error {
	src, err := strconv.Unquote(string(v))
	if err != nil {
		return fmt.Errorf("Action: UnmarshalJSON: %w", err)
	}

	action, err := ParseAction(src)
	if err != nil {
		return fmt.Errorf("Action: UnmarshalJSON: %w", err)
	}

	*a = action

	return nil
}

// String returns string representation of action.
func (a Action) String() string {
	if a.uid != "" {
		return a.uid
	}

	return common.Und
}

func (a Action) GoString() string {
	return "domain.Action(" + a.String() + ")"
}
