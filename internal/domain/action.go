package domain

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// NOTE(toby3d): Encapsulate enums in structs for extra compile-time safety:
// https://threedots.tech/post/safer-enums-in-go/#struct-based-enums
type Action struct {
	slug string
}

//nolint: gochecknoglobals // NOTE(toby3d): structs cannot be constants
var (
	ActionUndefined = Action{slug: ""}
	ActionRevoke    = Action{slug: "revoke"}
	ActionTicket    = Action{slug: "ticket"}
)

var ErrActionUnknown = errors.New("unknown action method")

// ParseAction parse string identifier of action into struct enum.
func ParseAction(slug string) (Action, error) {
	switch strings.ToLower(slug) {
	case ActionRevoke.slug:
		return ActionRevoke, nil
	case ActionTicket.slug:
		return ActionTicket, nil
	}

	return ActionUndefined, fmt.Errorf("%w: %s", ErrActionUnknown, slug)
}

// UnmarshalForm implements custom unmarshler for form values.
func (a *Action) UnmarshalForm(v []byte) error {
	action, err := ParseAction(string(v))
	if err != nil {
		return fmt.Errorf("action: %w", err)
	}

	*a = action

	return nil
}

func (a *Action) UnmarshalJSON(v []byte) error {
	src, err := strconv.Unquote(string(v))
	if err != nil {
		return err
	}

	action, err := ParseAction(src)
	if err != nil {
		return fmt.Errorf("action: %w", err)
	}

	*a = action

	return nil
}

// String returns string representation of action.
func (a Action) String() string {
	return a.slug
}
