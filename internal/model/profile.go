package model

type Profile struct {
	Name  string `json:"name"`
	URL   string `json:"url"`
	Photo string `json:"photo"`
	Email string `json:"email,omitempty"`
}
