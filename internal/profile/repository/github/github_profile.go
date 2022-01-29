package github

import (
	"context"
	"fmt"

	github "github.com/google/go-github/v41/github"
	"golang.org/x/oauth2"

	"source.toby3d.me/website/indieauth/internal/domain"
	"source.toby3d.me/website/indieauth/internal/profile"
)

type githubProfileRepository struct{}

const ErrPrefix string = "github"

func NewGithubProfileRepository() profile.Repository {
	return &githubProfileRepository{}
}

func (repo *githubProfileRepository) Get(ctx context.Context, token *oauth2.Token) (*domain.Profile, error) {
	user, _, err := github.NewClient(oauth2.NewClient(ctx, oauth2.StaticTokenSource(token))).Users.Get(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("%s: cannot get user info: %w", ErrPrefix, err)
	}

	result := new(domain.Profile)

	// NOTE(toby3d): Profile names.
	if user.Name != nil {
		result.Name = []string{*user.Name}
	}

	// NOTE(toby3d): Profile photos.
	if user.AvatarURL != nil {
		if u, err := domain.ParseURL(*user.AvatarURL); err == nil {
			result.Photo = []*domain.URL{u}
		}
	}

	// NOTE(toby3d): Profile URLs.
	result.URL = make([]*domain.URL, 0)
	var twitterURL *string

	if user.TwitterUsername != nil {
		u := "https://twitter.com/" + *user.TwitterUsername
		twitterURL = &u
	}

	for _, src := range []*string{
		user.HTMLURL, // NOTE(toby3d): always available.
		user.Blog,
		twitterURL,
	} {
		if src == nil {
			continue
		}

		u, err := domain.ParseURL(*src)
		if err != nil {
			continue
		}

		result.URL = append(result.URL, u)
	}

	// NOTE(toby3d): Profile Emails.
	if user.Email != nil {
		if email, err := domain.ParseEmail(*user.Email); err == nil {
			result.Email = []*domain.Email{email}
		}
	}

	return result, nil
}
