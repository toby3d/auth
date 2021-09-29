package domain_test

import (
	"encoding/base64"
	"io"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"source.toby3d.me/website/oauth/internal/domain"
	"source.toby3d.me/website/oauth/internal/random"
)

const (
	MinLength int = 42
	MaxLength int = 128
)

func TestPKCEIsValid(t *testing.T) {
	t.Parallel()

	rand.Seed(time.Now().UnixNano())

	//nolint: gosec
	verifier := random.New().String(MinLength + rand.Intn(MaxLength-MinLength))

	for _, testCase := range []struct {
		Name   string
		Method domain.PKCEMethod
	}{{
		Name:   "MD5",
		Method: domain.PKCEMethodMD5,
	}, {
		Name:   "plain",
		Method: domain.PKCEMethodPlain,
	}, {
		Name:   "S1",
		Method: domain.PKCEMethodS1,
	}, {
		Name:   "S256",
		Method: domain.PKCEMethodS256,
	}, {
		Name:   "S512",
		Method: domain.PKCEMethodS512,
	}, {
		Name:   "fallback to plain",
		Method: "UNDEFINED",
	}} {
		testCase := testCase

		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()

			pkce := &domain.PKCE{
				Method:    testCase.Method,
				Verifier:  verifier,
				Challenge: verifier,
			}

			if h := pkce.Method.Hash(); h != nil {
				_, err := io.WriteString(h, pkce.Verifier)
				require.NoError(t, err)

				pkce.Challenge = base64.RawURLEncoding.EncodeToString(h.Sum(nil))
			}

			assert.True(t, pkce.IsValid())
		})
	}
}
