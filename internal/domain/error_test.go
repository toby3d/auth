package domain_test

import (
	"fmt"
	"testing"

	"source.toby3d.me/website/indieauth/internal/domain"
)

func ExampleNewError() {
	fmt.Printf("%v", domain.NewError(domain.ErrorCodeInvalidRequest, "client_id MUST be provided", ""))
	// Output: invalid_request: client_id MUST be provided
}

func TestErrorCode_UnmarshalForm(t *testing.T) {
	t.Parallel()

	input := []byte("access_denied")
	result := domain.ErrorCodeUndefined

	if err := result.UnmarshalForm(input); err != nil {
		t.Fatalf("%+v", err)
	}

	if result != domain.ErrorCodeAccessDenied {
		t.Errorf("UnmarshalForm(%s) = %v, want %v", input, result, domain.ErrorCodeAccessDenied)
	}
}
