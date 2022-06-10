//nolint: dupl
package domain_test

import (
	"testing"

	"source.toby3d.me/toby3d/auth/internal/domain"
)

func TestResponseTypeType(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		in  string
		out domain.ResponseType
	}{
		{in: "id", out: domain.ResponseTypeID},
		{in: "code", out: domain.ResponseTypeCode},
	} {
		tc := tc

		t.Run(tc.in, func(t *testing.T) {
			t.Parallel()

			result, err := domain.ParseResponseType(tc.in)
			if err != nil {
				t.Fatalf("%+v", err)
			}

			if result != tc.out {
				t.Errorf("ParseResponseType(%s) = %v, want %v", tc.in, result, tc.out)
			}
		})
	}
}

func TestResponseType_UnmarshalForm(t *testing.T) {
	t.Parallel()

	input := []byte("code")
	result := domain.ResponseTypeUndefined

	if err := result.UnmarshalForm(input); err != nil {
		t.Fatalf("%+v", err)
	}

	if result != domain.ResponseTypeCode {
		t.Errorf("UnmarshalForm(%s) = %v, want %v", input, result, domain.ResponseTypeCode)
	}
}

func TestResponseType_UnmarshalJSON(t *testing.T) {
	t.Parallel()

	input := []byte(`"code"`)
	result := domain.ResponseTypeUndefined

	if err := result.UnmarshalJSON(input); err != nil {
		t.Fatalf("%+v", err)
	}

	if result != domain.ResponseTypeCode {
		t.Errorf("UnmarshalJSON(%s) = %v, want %v", input, result, domain.ResponseTypeCode)
	}
}

func TestResponseType_String(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		in   domain.ResponseType
		out  string
	}{
		{name: "id", in: domain.ResponseTypeID, out: "id"},
		{name: "code", in: domain.ResponseTypeCode, out: "code"},
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
