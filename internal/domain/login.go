package domain

import (
	"testing"
	"time"

	"source.toby3d.me/website/oauth/internal/random"
)

type Login struct {
	CreatedAt   time.Time
	CompletedAt time.Time
	PKCE
	Scopes      []string
	ClientID    string
	RedirectURI string
	MeEntered   string
	MeResolved  string
	Code        string
	Provider    string
	IsCompleted bool
}

//nolint: gomnd
func TestLogin(tb testing.TB) *Login {
	tb.Helper()

	now := time.Now().UTC()

	return &Login{
		CreatedAt:   now.Add(-1 * time.Minute),
		CompletedAt: time.Time{},
		PKCE: PKCE{
			Method:    PKCEMethodS256,
			Challenge: "OfYAxt8zU2dAPDWQxTAUIteRzMsoj9QBdMIVEDOErUo",
			Verifier:  "a6128783714cfda1d388e2e98b6ae8221ac31aca31959e59512c59f5",
		},
		Scopes:      []string{"profile", "create", "update", "delete"},
		ClientID:    "https://app.example.com/",
		RedirectURI: "https://app.example.com/redirect",
		MeEntered:   "user.example.net",
		MeResolved:  "https://user.example.net/",
		Code:        random.New().String(8),
		Provider:    "mastodon",
		IsCompleted: false,
	}
}

//nolint: gomnd
func TestLoginInvalid(tb testing.TB) *Login {
	tb.Helper()

	now := time.Now().UTC()

	return &Login{
		CreatedAt:   now.Add(-1 * time.Hour),
		CompletedAt: time.Time{},
		PKCE: PKCE{
			Method:    "UNDEFINED",
			Challenge: random.New().String(42),
			Verifier:  random.New().String(64),
		},
		Scopes:      []string{},
		ClientID:    "whoisit",
		RedirectURI: "redirect",
		MeEntered:   "whoami",
		MeResolved:  "",
		Code:        "",
		Provider:    "",
		IsCompleted: true,
	}
}
