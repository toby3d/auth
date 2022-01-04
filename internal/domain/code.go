//nolint: gosec
package domain

import (
	"encoding/base64"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/require"

	"source.toby3d.me/website/indieauth/internal/random"
)

// Code describes the PKCE challenge to validate the security of the request.
type Code struct {
	Method    CodeChallengeMethod
	Verifier  string
	Challenge string
}

const (
	CodeLengthMin int = 43
	CodeLengthMax int = 128
)

// TestCode returns valid random generated PKCE code for tests.
func TestCode(tb testing.TB) *Code {
	tb.Helper()

	verifier, err := random.String(
		gofakeit.Number(CodeLengthMin, CodeLengthMax), random.Alphanumeric, "-", ".", "_", "~",
	)
	require.NoError(tb, err)

	h := CodeChallengeMethodS256.hash
	h.Reset()

	_, err = h.Write([]byte(verifier))
	require.NoError(tb, err)

	return &Code{
		Method:    CodeChallengeMethodS256,
		Verifier:  verifier,
		Challenge: base64.RawURLEncoding.EncodeToString(h.Sum(nil)),
	}
}

// IsValid returns true if code challenge is equal to the generated hash from
// the verifier.
func (c Code) IsValid() bool {
	if c.Method == CodeChallengeMethodUndefined {
		return false
	}

	if c.Method == CodeChallengeMethodPLAIN {
		return c.Challenge == c.Verifier
	}

	h := c.Method.hash
	h.Reset()

	return c.Challenge == base64.RawURLEncoding.EncodeToString(h.Sum([]byte(c.Verifier)))
}
