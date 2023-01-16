package http

import "source.toby3d.me/toby3d/auth/internal/domain"

type UserInformationResponse struct {
	Name  string        `json:"name,omitempty"`
	URL   *domain.URL   `json:"url,omitempty"`
	Photo *domain.URL   `json:"photo,omitempty"`
	Email *domain.Email `json:"email,omitempty"`
}

func NewUserInformationResponse(in *domain.Profile, hasEmail bool) *UserInformationResponse {
	out := new(UserInformationResponse)

	if in == nil {
		return out
	}

	if in.HasName() {
		out.Name = in.GetName()
	}

	if in.HasURL() {
		out.URL = &domain.URL{URL: in.GetURL()}
	}

	if in.HasPhoto() {
		out.Photo = &domain.URL{URL: in.GetPhoto()}
	}

	if hasEmail {
		out.Email = in.GetEmail()
	}

	return out
}
