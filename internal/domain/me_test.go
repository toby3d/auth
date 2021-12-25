package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"source.toby3d.me/website/oauth/internal/domain"
)

func TestMe(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name    string
		input   string
		isValid bool
	}{{
		name:    "valid",
		input:   "https://example.com/",
		isValid: true,
	}, {
		name:    "valid with path",
		input:   "https://example.com/username",
		isValid: true,
	}, {
		name:    "valid with query",
		input:   "https://example.com/users?id=100",
		isValid: true,
	}, {
		name:    "missing scheme",
		input:   "example.com",
		isValid: false,
	}, {
		name:    "invalid scheme",
		input:   "mailto:user@example.com",
		isValid: false,
	}, {
		name:    "contains a double-dot path segment",
		input:   "https://example.com/foo/../bar",
		isValid: false,
	}, {
		name:    "contains a fragment",
		input:   "https://example.com/#me",
		isValid: false,
	}, {
		name:    "contains a username and password",
		input:   "https://user:pass@example.com/",
		isValid: false,
	}, {
		name:    "contains a port",
		input:   "https://example.com:8443/",
		isValid: false,
	}, {
		name:    "host is an IP address",
		input:   "https://172.28.92.51/",
		isValid: false,
	}} {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result, err := domain.NewMe(testCase.input)
			if testCase.isValid {
				require.NoError(t, err)
				assert.Equal(t, testCase.input, result.String())
			} else {
				assert.Error(t, err)
			}
		})
	}
}
