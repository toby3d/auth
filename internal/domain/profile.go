package domain

import (
	"net/url"
	"testing"
)

// Profile describes the data about the user.
type Profile struct {
	Photo *url.URL `json:"photo,omitempty"`
	URL   *url.URL `json:"url,omitempty"`
	Email *Email   `json:"email,omitempty"`
	Name  string   `json:"name,omitempty"`
}

func NewProfile() *Profile {
	return &Profile{
		Photo: new(url.URL),
		URL:   new(url.URL),
		Email: new(Email),
		Name:  "",
	}
}

// TestProfile returns a valid Profile with the generated test data filled in.
func TestProfile(tb testing.TB) *Profile {
	tb.Helper()

	return &Profile{
		Email: TestEmail(tb),
		Name:  "Example User",
		Photo: &url.URL{Scheme: "https", Host: "user.example.net", Path: "/photo.jpg"},
		URL:   &url.URL{Scheme: "https", Host: "user.example.net", Path: "/"},
	}
}
