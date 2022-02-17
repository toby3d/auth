package domain

import (
	"testing"

	"source.toby3d.me/website/indieauth/internal/random"
)

//nolint: tagliatelle
type Session struct {
	ClientID            *ClientID           `json:"client_id"`
	RedirectURI         *URL                `json:"redirect_uri"`
	Me                  *Me                 `json:"me"`
	Profile             *Profile            `json:"profile,omitempty"`
	Scope               Scopes              `json:"scope"`
	CodeChallengeMethod CodeChallengeMethod `json:"code_challenge_method,omitempty"`
	CodeChallenge       string              `json:"code_challenge,omitempty"`
	Code                string              `json:"-"`
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
		Profile:             TestProfile(tb),
		Me:                  TestMe(tb, "https://user.example.net/"),
		RedirectURI:         TestURL(tb, "https://example.com/callback"),
		Scope: Scopes{
			ScopeEmail,
			ScopeProfile,
		},
	}
}
