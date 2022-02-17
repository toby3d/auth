package domain

import (
	"testing"
)

// Profile describes the data about the user.
type Profile struct {
	Photo []*URL   `json:"photo"`
	URL   []*URL   `json:"url"`
	Email []*Email `json:"email"`
	Name  []string `json:"name"`
}

func NewProfile() *Profile {
	return &Profile{
		Photo: make([]*URL, 0),
		URL:   make([]*URL, 0),
		Email: make([]*Email, 0),
		Name:  make([]string, 0),
	}
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

// GetName safe returns first name, if any.
func (p Profile) GetName() string {
	if len(p.Name) == 0 {
		return ""
	}

	return p.Name[0]
}

// GetURL safe returns first URL, if any.
func (p Profile) GetURL() *URL {
	if len(p.URL) == 0 {
		return nil
	}

	return p.URL[0]
}

// GetPhoto safe returns first photo, if any.
func (p Profile) GetPhoto() *URL {
	if len(p.Photo) == 0 {
		return nil
	}

	return p.Photo[0]
}

// GetEmail safe returns first email, if any.
func (p Profile) GetEmail() *Email {
	if len(p.Email) == 0 {
		return nil
	}

	return p.Email[0]
}
