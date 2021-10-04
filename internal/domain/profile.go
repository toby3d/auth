package domain

import "testing"

type Profile struct {
	Name  string
	URL   string
	Photo string
	Email string
}

func TestProfile(tb testing.TB) *Profile {
	tb.Helper()

	return &Profile{
		Name:  "Example User",
		URL:   "https://user.example.net/",
		Photo: "https://user.example.net/photo.jpg",
		Email: "user@example.net",
	}
}
