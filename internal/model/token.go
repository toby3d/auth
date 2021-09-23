package model

type Token struct {
	Profile     *Profile
	Scopes      []string
	AccessToken string
	Type        string
	Me          URL
	ClientID    URL
}

func NewToken() *Token {
	t := new(Token)
	t.Scopes = make([]string, 0)

	return t
}
