package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"source.toby3d.me/website/oauth/internal/domain"
)

func TestParseCodeChallengeMethod(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		output   domain.CodeChallengeMethod
		name     string
		input    string
		expError bool
	}{{
		expError: true,
		name:     "invalid",
		input:    "und",
		output:   domain.CodeChallengeMethodUndefined,
	}, {
		name:   "PLAIN",
		input:  "plain",
		output: domain.CodeChallengeMethodPLAIN,
	}, {
		name:   "MD5",
		input:  "Md5",
		output: domain.CodeChallengeMethodMD5,
	}, {
		name:   "S1",
		input:  "S1",
		output: domain.CodeChallengeMethodS1,
	}, {
		name:   "S256",
		input:  "S256",
		output: domain.CodeChallengeMethodS256,
	}, {
		name:   "S512",
		input:  "S512",
		output: domain.CodeChallengeMethodS512,
	}} {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result, err := domain.ParseCodeChallengeMethod(testCase.input)
			if testCase.expError {
				assert.Error(t, err)
				assert.Equal(t, domain.CodeChallengeMethodUndefined, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.output, result)
			}
		})
	}
}
