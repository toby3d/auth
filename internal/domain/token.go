package domain

import (
	"testing"

	"source.toby3d.me/website/oauth/internal/random"
)

type Token struct {
	AccessToken string
	ClientID    string
	Me          string
	Profile     *Profile
	Scopes      []string
	Type        string
}

func NewToken() *Token {
	t := new(Token)
	t.Scopes = make([]string, 0)

	return t
}

func TestToken(tb testing.TB) *Token {
	tb.Helper()

	return &Token{
		AccessToken: random.New().String(32),
		ClientID:    "https://app.example.com/",
		Me:          "https://user.example.net/",
		Profile:     TestProfile(tb),
		Scopes:      []string{"create", "update", "delete"},
		Type:        "Bearer",
	}
}
