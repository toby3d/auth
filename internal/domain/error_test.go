package domain_test

import (
	"fmt"

	"source.toby3d.me/website/indieauth/internal/domain"
)

func ExampleNewError() {
	fmt.Printf("%v", domain.NewError(domain.ErrorCodeInvalidRequest, "client_id MUST be provided", ""))
	// Output: invalid_request: client_id MUST be provided
}
