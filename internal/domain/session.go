package domain

import (
	"testing"

	"github.com/stretchr/testify/require"

	"source.toby3d.me/website/indieauth/internal/random"
)

type Session struct {
	ClientID            *ClientID
	Me                  *Me
	RedirectURI         *URL
	Profile             *Profile
	CodeChallengeMethod CodeChallengeMethod
	Scope               Scopes
	Code                string
	CodeChallenge       string
}

// TestSession returns valid random generated session for tests.
func TestSession(tb testing.TB) *Session {
	tb.Helper()

	code, err := random.String(24)
	require.NoError(tb, err)

	return &Session{
		ClientID:            TestClientID(tb),
		Code:                code,
		CodeChallenge:       "hackme",
		CodeChallengeMethod: CodeChallengeMethodPLAIN,
		Me:                  TestMe(tb, "https://user.example.net/"),
		RedirectURI:         TestURL(tb, "https://example.com/callback"),
		Scope: Scopes{
			ScopeEmail,
			ScopeProfile,
		},
	}
}
