package http

import "source.toby3d.me/toby3d/auth/internal/domain"

type UserInformationResponse struct {
	URL   string `json:"url,omitempty"`
	Photo string `json:"photo,omitempty"`
	Email string `json:"email,omitempty"`
	Name  string `json:"name,omitempty"`
}

func NewUserInformationResponse(in *domain.Profile, hasEmail bool) *UserInformationResponse {
	out := new(UserInformationResponse)

	if in == nil {
		return out
	}

	out.Name = in.Name

	if in.URL != nil {
		out.URL = in.URL.String()
	}

	if in.Photo != nil {
		out.Photo = in.Photo.String()
	}

	if hasEmail && in.Email != nil {
		out.Email = in.Email.String()
	}

	return out
}
