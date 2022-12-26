//nolint:dupl
package domain_test

import (
	"testing"

	"source.toby3d.me/toby3d/auth/internal/domain"
)

func TestParseGrantType(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		in  string
		out domain.GrantType
	}{
		{in: "authorization_code", out: domain.GrantTypeAuthorizationCode},
		{in: "ticket", out: domain.GrantTypeTicket},
	} {
		tc := tc

		t.Run(tc.in, func(t *testing.T) {
			t.Parallel()

			result, err := domain.ParseGrantType(tc.in)
			if err != nil {
				t.Fatalf("%+v", err)
			}

			if result != tc.out {
				t.Errorf("ParseGrantType(%s) = %v, want %v", tc.in, result, tc.out)
			}
		})
	}
}

func TestGrantType_UnmarshalForm(t *testing.T) {
	t.Parallel()

	input := []byte("authorization_code")
	result := domain.GrantTypeUnd

	if err := result.UnmarshalForm(input); err != nil {
		t.Fatalf("%+v", err)
	}

	if result != domain.GrantTypeAuthorizationCode {
		t.Errorf("UnmarshalForm(%s) = %v, want %v", input, result, domain.GrantTypeAuthorizationCode)
	}
}

func TestGrantType_UnmarshalJSON(t *testing.T) {
	t.Parallel()

	input := []byte(`"authorization_code"`)
	result := domain.GrantTypeUnd

	if err := result.UnmarshalJSON(input); err != nil {
		t.Fatalf("%+v", err)
	}

	if result != domain.GrantTypeAuthorizationCode {
		t.Errorf("UnmarshalJSON(%s) = %v, want %v", input, result, domain.GrantTypeAuthorizationCode)
	}
}

func TestGrantType_String(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		in   domain.GrantType
		out  string
	}{
		{name: "authorization_code", in: domain.GrantTypeAuthorizationCode, out: "authorization_code"},
		{name: "ticket", in: domain.GrantTypeTicket, out: "ticket"},
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
