package domain

import (
	"testing"

	"source.toby3d.me/website/indieauth/internal/random"
)

type Session struct {
	ClientID            *ClientID
	RedirectURI         *URL
	Me                  *Me
	CodeChallengeMethod CodeChallengeMethod
	Scope               Scopes
	CodeChallenge       string
	Code                string
}

// TestSession returns valid random generated session for tests.
//nolint: gomnd // testing domain can contains non-standart values
func TestSession(tb testing.TB) *Session {
	tb.Helper()

	code, err := random.String(24)
	if err != nil {
		tb.Fatal(err)
	}

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
