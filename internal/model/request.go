package model

type ExchangeRequest struct {
	Code         string
	ClientID     string
	RedirectURI  string
	CodeVerifier string
}
