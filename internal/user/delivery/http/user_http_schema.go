package http

type UserInformationResponse struct {
	Name  string `json:"name,omitempty"`
	URL   string `json:"url,omitempty"`
	Photo string `json:"photo,omitempty"`
	Email string `json:"email,omitempty"`
}
