package domain

import (
	"fmt"
	"strconv"
	"strings"
)

// Scopes represent set of Scope domains.
type Scopes []Scope

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
		if s[i] == ScopeUnd {
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
