package domain

import (
	"testing"
)

// Profile describes the data about the user.
type Profile struct {
	Photo []*URL   `json:"photo,omitempty"`
	URL   []*URL   `json:"url,omitempty"`
	Email []*Email `json:"email,omitempty"`
	Name  []string `json:"name,omitempty"`
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

func (p Profile) HasName() bool {
	return len(p.Name) > 0
}

// GetName safe returns first name, if any.
func (p Profile) GetName() string {
	if len(p.Name) == 0 {
		return ""
	}

	return p.Name[0]
}

func (p Profile) HasURL() bool {
	return len(p.URL) > 0
}

// GetURL safe returns first URL, if any.
func (p Profile) GetURL() *URL {
	if len(p.URL) == 0 {
		return nil
	}

	return p.URL[0]
}

func (p Profile) HasPhoto() bool {
	return len(p.Photo) > 0
}

// GetPhoto safe returns first photo, if any.
func (p Profile) GetPhoto() *URL {
	if len(p.Photo) == 0 {
		return nil
	}

	return p.Photo[0]
}

func (p Profile) HasEmail() bool {
	return len(p.Email) > 0
}

// GetEmail safe returns first email, if any.
func (p Profile) GetEmail() *Email {
	if len(p.Email) == 0 {
		return nil
	}

	return p.Email[0]
}
