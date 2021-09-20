package model

type Token struct {
	AccessToken string
	ClientID    string
	Me          string
	Profile     *Profile
	Scopes      []string
	Type        string
}
