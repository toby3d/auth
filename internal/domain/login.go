package domain

import (
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/require"

	"source.toby3d.me/website/oauth/internal/random"
)

type Login struct {
	PKCE
	CreatedAt   time.Time
	CompletedAt time.Time
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

	code, err := random.String(8)
	require.NoError(tb, err)

	return &Login{
		CreatedAt:   gofakeit.Date(),
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
		Code:        code,
		Provider:    "mastodon",
		IsCompleted: false,
	}
}

//nolint: gomnd
func TestLoginInvalid(tb testing.TB) *Login {
	tb.Helper()

	challenge, err := random.String(42)
	require.NoError(tb, err)

	verifier, err := random.String(64)
	require.NoError(tb, err)

	return &Login{
		CreatedAt:   time.Now().UTC().Add(-1 * time.Hour),
		CompletedAt: time.Time{},
		PKCE: PKCE{
			Method:    "UNDEFINED",
			Challenge: challenge,
			Verifier:  verifier,
		},
		Scopes:      []string{},
		ClientID:    "whoisit",
		RedirectURI: "redirect",
		MeEntered:   "whoami",
		MeResolved:  "",
		Code:        "",
		Provider:    "undefined",
		IsCompleted: true,
	}
}
