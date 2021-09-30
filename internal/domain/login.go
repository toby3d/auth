package domain

import (
	"testing"

	"source.toby3d.me/website/oauth/internal/random"
)

type Login struct {
	ClientID            string
	Code                string
	CodeChallenge       string
	CodeChallengeMethod PKCEMethod
	CodeVerifier        string
	Me                  string
	RedirectURI         string
	Scopes              []string
	State               string
}

//nolint: gomnd
func TestLogin(tb testing.TB) *Login {
	tb.Helper()

	return &Login{
		ClientID:            "http://app.example.com/",
		Code:                random.New().String(16),
		CodeChallenge:       "OfYAxt8zU2dAPDWQxTAUIteRzMsoj9QBdMIVEDOErUo",
		CodeChallengeMethod: PKCEMethodS256,
		CodeVerifier:        "a6128783714cfda1d388e2e98b6ae8221ac31aca31959e59512c59f5",
		Me:                  "http://user.example.net/",
		RedirectURI:         "http://app.example.com/redirect",
		Scopes:              []string{"profile", "create", "update", "delete"},
		State:               "1234567890",
	}
}

//nolint: gomnd
func TestLoginInvalid(tb testing.TB) *Login {
	tb.Helper()

	return &Login{
		ClientID:            "whoisit",
		Code:                "",
		CodeChallenge:       random.New().String(42),
		CodeChallengeMethod: "UNDEFINED",
		CodeVerifier:        random.New().String(42),
		Me:                  "whoami",
		RedirectURI:         "/redirect",
		Scopes:              []string{},
		State:               "",
	}
}
