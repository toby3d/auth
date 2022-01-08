package domain_test

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"hash"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"source.toby3d.me/website/indieauth/internal/domain"
	"source.toby3d.me/website/indieauth/internal/random"
)

func TestCodeChallengeMethod_Parse(t *testing.T) {
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

//nolint: funlen
func TestCodeChallengeMethod_Validate(t *testing.T) {
	t.Parallel()

	verifier, err := random.String(gofakeit.Number(42, 128))
	require.NoError(t, err)

	for _, testCase := range []struct {
		hash    hash.Hash
		name    string
		method  string
		isValid bool
	}{{
		name:    "invalid",
		method:  domain.CodeChallengeMethodS256.String(),
		hash:    md5.New(),
		isValid: false,
	}, {
		name:    "MD5",
		method:  domain.CodeChallengeMethodMD5.String(),
		hash:    md5.New(),
		isValid: true,
	}, {
		name:    "plain",
		method:  domain.CodeChallengeMethodPLAIN.String(),
		hash:    nil,
		isValid: true,
	}, {
		name:    "S1",
		method:  domain.CodeChallengeMethodS1.String(),
		hash:    sha1.New(),
		isValid: true,
	}, {
		name:    "S256",
		method:  domain.CodeChallengeMethodS256.String(),
		hash:    sha256.New(),
		isValid: true,
	}, {
		name:    "S512",
		method:  domain.CodeChallengeMethodS512.String(),
		hash:    sha512.New(),
		isValid: true,
	}, {
		name:    "undefined",
		method:  "und",
		hash:    nil,
		isValid: false,
	}} {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			method, _ := domain.ParseCodeChallengeMethod(testCase.method)
			if method == domain.CodeChallengeMethodPLAIN ||
				method == domain.CodeChallengeMethodUndefined {
				assert.Equal(t, testCase.isValid, method.Validate(verifier, verifier))

				return
			}

			assert.Equal(t, testCase.isValid, method.Validate(base64.RawURLEncoding.EncodeToString(
				testCase.hash.Sum([]byte(verifier)),
			), verifier))
		})
	}
}
