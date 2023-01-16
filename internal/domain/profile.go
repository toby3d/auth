package domain

import (
	"net/url"
	"testing"
)

// Profile describes the data about the user.
type Profile struct {
	Photo []*url.URL `json:"photo,omitempty"`
	URL   []*url.URL `json:"url,omitempty"`
	Email []*Email   `json:"email,omitempty"`
	Name  []string   `json:"name,omitempty"`
}

func NewProfile() *Profile {
	return &Profile{
		Photo: make([]*url.URL, 0),
		URL:   make([]*url.URL, 0),
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
		Photo: []*url.URL{{Scheme: "https", Host: "user.example.net", Path: "/photo.jpg"}},
		URL:   []*url.URL{{Scheme: "https", Host: "user.example.net", Path: "/"}},
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
func (p Profile) GetURL() *url.URL {
	if len(p.URL) == 0 {
		return nil
	}

	return p.URL[0]
}

func (p Profile) HasPhoto() bool {
	return len(p.Photo) > 0
}

// GetPhoto safe returns first photo, if any.
func (p Profile) GetPhoto() *url.URL {
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
