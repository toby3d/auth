package domain

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Profile describes the data about the user.
type Profile struct {
	Photo []*URL
	URL   []*URL
	Email []Email
	Name  []string
}

// NewProfile creates a new empty Profile.
func NewProfile() *Profile {
	return &Profile{
		Email: make([]Email, 0),
		Name:  make([]string, 0),
		Photo: make([]*URL, 0),
		URL:   make([]*URL, 0),
	}
}

// TestProfile returns a valid Profile with the generated test data filled in.
func TestProfile(tb testing.TB) *Profile {
	tb.Helper()

	photo, err := NewURL("https://user.example.net/photo.jpg")
	require.NoError(tb, err)

	url, err := NewURL("https://user.example.net/")
	require.NoError(tb, err)

	return &Profile{
		Email: []Email{"user@example.net"},
		Name:  []string{"Example User"},
		Photo: []*URL{photo},
		URL:   []*URL{url},
	}
}
