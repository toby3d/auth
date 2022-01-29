package domain

import "testing"

type Session struct {
	ClientID            *ClientID
	Me                  *Me
	RedirectURI         *URL
	CodeChallengeMethod CodeChallengeMethod
	Scope               Scopes
	Code                string
	CodeChallenge       string
}

func TestSession(tb testing.TB) *Session {
	tb.Helper()

	return &Session{
		ClientID:            TestClientID(tb),
		Me:                  TestMe(tb, "https://user.example.net/"),
		RedirectURI:         TestURL(tb, "https://example.com/callback"),
		CodeChallengeMethod: CodeChallengeMethodPLAIN,
		Scope:               Scopes{ScopeProfile, ScopeEmail},
		Code:                "abcdefg",
		CodeChallenge:       "hackme",
	}
}
