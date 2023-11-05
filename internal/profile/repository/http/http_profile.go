package http

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"golang.org/x/exp/slices"
	"willnorris.com/go/microformats"

	"source.toby3d.me/toby3d/auth/internal/common"
	"source.toby3d.me/toby3d/auth/internal/domain"
	"source.toby3d.me/toby3d/auth/internal/profile"
)

type httpProfileRepository struct {
	client *http.Client
}

func NewHTPPClientRepository(client *http.Client) profile.Repository {
	return &httpProfileRepository{
		client: client,
	}
}

// WARN(toby3d): not implemented.
func (repo *httpProfileRepository) Create(_ context.Context, _ domain.Me, _ domain.Profile) error {
	return nil
}

//nolint:cyclop,funlen
func (repo *httpProfileRepository) Get(_ context.Context, me domain.Me) (*domain.Profile, error) {
	resp, err := repo.client.Get(me.String())
	if err != nil {
		return nil, fmt.Errorf("%w: cannot fetch user by me: %w", profile.ErrNotExist, err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: cannot read response body: %w", profile.ErrNotExist, err)
	}

	mf2 := microformats.Parse(bytes.NewReader(body), resp.Request.URL)
	out := new(domain.Profile)

	for i := range mf2.Items {
		if !slices.Contains(mf2.Items[i].Type, common.HCard) {
			continue
		}

		parseProfile(mf2.Items[i].Properties, out)
	}

	return out, nil
}

func parseProfile(src map[string][]any, dst *domain.Profile) {
	for _, val := range src[common.PropertyName] {
		v, ok := val.(string)
		if !ok {
			continue
		}

		dst.Name = v

		break
	}

	for _, val := range src[common.PropertyURL] {
		v, ok := val.(string)
		if !ok {
			continue
		}

		var err error
		if dst.URL, err = url.Parse(v); err != nil {
			continue
		}

		break
	}

	for _, val := range src[common.PropertyPhoto] {
		v, ok := val.(string)
		if !ok {
			continue
		}

		var err error
		if dst.Photo, err = url.Parse(v); err != nil {
			continue
		}

		break
	}
}
