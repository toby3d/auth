package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"source.toby3d.me/website/indieauth/internal/domain"
)

func TestParseEmail(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name      string
		input     string
		expError  bool
		expResult string
	}{{
		name:      "simple",
		input:     "user@example.com",
		expError:  false,
		expResult: "user@example.com",
	}, {
		name:      "subaddress",
		input:     "user+suffix@example.com",
		expError:  false,
		expResult: "user+suffix@example.com",
	}, {
		name:      "prefix",
		input:     "mailto:user@example.com",
		expError:  false,
		expResult: "user@example.com",
	}} {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result, err := domain.ParseEmail(testCase.input)
			if testCase.expError {
				assert.Error(t, err)
				assert.Nil(t, result)

				return
			}

			assert.NoError(t, err)
			assert.Equal(t, testCase.expResult, result.String())
		})
	}
}
