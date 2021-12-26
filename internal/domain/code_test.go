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

	"source.toby3d.me/website/oauth/internal/domain"
	"source.toby3d.me/website/oauth/internal/random"
)

//nolint: funlen
func TestCodeIsValid(t *testing.T) {
	t.Parallel()

	verifier, err := random.String(gofakeit.Number(domain.CodeLengthMin, domain.CodeLengthMax))
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
			result := &domain.Code{
				Method:    method,
				Verifier:  verifier,
				Challenge: verifier,
			}

			if method == domain.CodeChallengeMethodPLAIN ||
				method == domain.CodeChallengeMethodUndefined {
				assert.Equal(t, testCase.isValid, result.IsValid())

				return
			}

			result.Challenge = base64.RawURLEncoding.EncodeToString(
				testCase.hash.Sum([]byte(result.Verifier)),
			)
			assert.Equal(t, testCase.isValid, result.IsValid())
		})
	}
}
