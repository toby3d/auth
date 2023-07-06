package domain

import (
	"net/url"
	"testing"

	"source.toby3d.me/toby3d/auth/internal/random"
)

//nolint:tagliatelle
type Session struct {
	ClientID            ClientID            `json:"client_id"`
	RedirectURI         *url.URL            `json:"redirect_uri"`
	Me                  Me                  `json:"me"`
	Profile             *Profile            `json:"profile,omitempty"`
	CodeChallengeMethod CodeChallengeMethod `json:"code_challenge_method,omitempty"`
	CodeChallenge       string              `json:"code_challenge,omitempty"`
	Code                string              `json:"-"`
	Scope               Scopes              `json:"scope"`
}

// TestSession returns valid random generated session for tests.
//
//nolint:gomnd // testing domain can contains non-standart values
func TestSession(tb testing.TB) *Session {
	tb.Helper()

	code, err := random.String(24)
	if err != nil {
		tb.Fatal(err)
	}

	return &Session{
		ClientID:            *TestClientID(tb),
		Code:                code,
		CodeChallenge:       "hackme",
		CodeChallengeMethod: CodeChallengeMethodPLAIN,
		Profile:             TestProfile(tb),
		Me:                  *TestMe(tb, "https://user.example.net/"),
		RedirectURI:         &url.URL{Scheme: "https", Host: "example.com", Path: "/callback"},
		Scope: Scopes{
			ScopeEmail,
			ScopeProfile,
		},
	}
}
