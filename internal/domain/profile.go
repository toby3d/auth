package domain

import (
	"testing"
)

// Profile describes the data about the user.
type Profile struct {
	Photo []*URL
	URL   []*URL
	Email []*Email
	Name  []string
}

// TestProfile returns a valid Profile with the generated test data filled in.
func TestProfile(tb testing.TB) *Profile {
	tb.Helper()

	return &Profile{
		Email: []*Email{TestEmail(tb)},
		Name:  []string{"Example User"},
		Photo: []*URL{TestURL(tb, "https://user.example.net/photo.jpg")},
		URL:   []*URL{TestURL(tb, "https://user.example.net/")},
	}
}
