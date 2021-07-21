package model

type Login struct {
	CreatedAt           int64  `json:"created_at"`
	ClientID            string `json:"client_id"`
	Code                string `json:"code"`
	CodeChallenge       string `json:"code_challenge"`
	CodeChallengeMethod string `json:"code_challenge_method"`
	Me                  string `json:"me"`
	RedirectURI         string `json:"redirect_uri"`
	Scope               string `json:"scope"`
}

func (Login) Bucket() []byte { return []byte("logins") }
