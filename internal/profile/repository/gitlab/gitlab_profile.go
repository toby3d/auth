package gitlab

import (
	"context"
	"fmt"

	gitlab "github.com/xanzy/go-gitlab"
	"golang.org/x/oauth2"

	"source.toby3d.me/website/indieauth/internal/domain"
	"source.toby3d.me/website/indieauth/internal/profile"
)

type gitlabProfileRepository struct{}

const ErrPrefix string = "gitlab"

func NewGitlabProfileRepository() profile.Repository {
	return &gitlabProfileRepository{}
}

//nolint: funlen // NOTE(toby3d): uses hyphenation on new lines for readability.
func (repo *gitlabProfileRepository) Get(_ context.Context, token *oauth2.Token) (*domain.Profile, error) {
	client, err := gitlab.NewClient(token.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("%s: cannot create client: %w", ErrPrefix, err)
	}

	user, _, err := client.Users.CurrentUser()
	if err != nil {
		return nil, fmt.Errorf("%s: cannot get user info: %w", ErrPrefix, err)
	}

	result := new(domain.Profile)

	// NOTE(toby3d): Profile names.
	if user.Name != "" {
		result.Name = []string{user.Name}
	}

	// NOTE(toby3d): Profile photos.
	if user.AvatarURL != "" {
		if u, err := domain.ParseURL(user.AvatarURL); err == nil {
			result.Photo = []*domain.URL{u}
		}
	}

	// NOTE(toby3d): Profile URLs.
	result.URL = make([]*domain.URL, 0)

	for _, src := range []string{
		user.WebURL, // NOTE(toby3d): always available.
		user.WebsiteURL,
		"https://twitter.com/" + user.Twitter,
		// TODO(toby3d): Skype field
		// TODO(toby3d): LinkedIn field
	} {
		if src == "" || src == "https://twitter.com/" {
			continue
		}

		u, err := domain.ParseURL(user.WebsiteURL)
		if err != nil {
			continue
		}

		result.URL = append(result.URL, u)
	}

	// NOTE(toby3d): Profile Emails.
	result.Email = make([]*domain.Email, 0)

	for _, src := range []string{
		user.PublicEmail,
		user.Email,
	} {
		if src == "" {
			continue
		}

		email, err := domain.ParseEmail(src)
		if err != nil {
			continue
		}

		result.Email = append(result.Email, email)
	}

	return result, nil
}
