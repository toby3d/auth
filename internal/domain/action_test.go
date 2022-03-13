//nolint: dupl
package domain_test

import (
	"testing"

	"source.toby3d.me/toby3d/auth/internal/domain"
)

func TestParseAction(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		in  string
		out domain.Action
	}{
		{in: "revoke", out: domain.ActionRevoke},
		{in: "ticket", out: domain.ActionTicket},
	} {
		tc := tc

		t.Run(tc.in, func(t *testing.T) {
			t.Parallel()

			result, err := domain.ParseAction(tc.in)
			if err != nil {
				t.Fatalf("%+v", err)
			}

			if result != tc.out {
				t.Errorf("ParseAction(%s) = %v, want %v", tc.in, result, tc.out)
			}
		})
	}
}

func TestAction_UnmarshalForm(t *testing.T) {
	t.Parallel()

	input := []byte("revoke")
	result := domain.ActionUndefined

	if err := result.UnmarshalForm(input); err != nil {
		t.Fatalf("%+v", err)
	}

	if result != domain.ActionRevoke {
		t.Errorf("UnmarshalForm(%s) = %v, want %v", input, result, domain.ActionRevoke)
	}
}

func TestAction_UnmarshalJSON(t *testing.T) {
	t.Parallel()

	input := []byte(`"revoke"`)
	result := domain.ActionUndefined

	if err := result.UnmarshalJSON(input); err != nil {
		t.Fatalf("%+v", err)
	}

	if result != domain.ActionRevoke {
		t.Errorf("UnmarshalJSON(%s) = %v, want %v", input, result, domain.ActionRevoke)
	}
}

func TestAction_String(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		in   domain.Action
		out  string
	}{
		{name: "revoke", in: domain.ActionRevoke, out: "revoke"},
		{name: "ticket", in: domain.ActionTicket, out: "ticket"},
	} {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := tc.in.String()
			if result != tc.out {
				t.Errorf("String() = %v, want %v", result, tc.out)
			}
		})
	}
}
