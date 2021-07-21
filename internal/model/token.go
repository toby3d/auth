package model

type Token struct {
	AccessToken string   `json:"access_token"`
	TokenType   string   `json:"token_type"`
	Scope       string   `json:"scope"`
	Me          string   `json:"me"`
	ClientID    string   `json:"client_id"`
	Profile     *Profile `json:"profile,omitempty"`
}

func (Token) Bucket() []byte { return []byte("tokens") }
