package mastodon

import (
	"context"
	"fmt"

	mastodon "github.com/mattn/go-mastodon"
	"golang.org/x/oauth2"

	"source.toby3d.me/website/indieauth/internal/domain"
	"source.toby3d.me/website/indieauth/internal/profile"
)

type mastodonProfileRepository struct {
	server string
}

const ErrPrefix string = "mastodon"

func NewMastodonProfileRepository(server string) profile.Repository {
	return &mastodonProfileRepository{
		server: server,
	}
}

func (repo *mastodonProfileRepository) Get(ctx context.Context, token *oauth2.Token) (*domain.Profile, error) {
	account, err := mastodon.NewClient(&mastodon.Config{
		Server:      repo.server,
		AccessToken: token.AccessToken,
	}).GetAccountCurrentUser(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: cannot get account info: %w", ErrPrefix, err)
	}

	result := new(domain.Profile)

	// NOTE(toby3d): Profile names.
	if account.DisplayName != "" {
		result.Name = []string{account.DisplayName}
	}

	// NOTE(toby3d): Profile photos.
	if account.Avatar != "" {
		if u, err := domain.ParseURL(account.Avatar); err == nil {
			result.Photo = []*domain.URL{u}
		}
	}

	// NOTE(toby3d): Profile URLs.
	result.URL = make([]*domain.URL, 0)

	// NOTE(toby3d): must be always available
	if account.URL != "" {
		if u, err := domain.ParseURL(account.URL); err == nil {
			result.URL = append(result.URL, u)
		}
	}

	for i := range account.Fields {
		// NOTE(toby3d): ignore non-verified fields that contain either
		// free-form text or links in them have not yet been verified.
		if account.Fields[i].VerifiedAt.IsZero() {
			continue
		}

		u, err := domain.ParseURL(account.Fields[i].Value)
		if err != nil {
			continue
		}

		result.URL = append(result.URL, u)
	}

	// WARN(toby3d): Mastodon does not provide an email via API.

	return result, nil
}
