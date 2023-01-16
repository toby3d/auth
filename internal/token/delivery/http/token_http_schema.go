package http

import (
	"errors"
	"net/http"

	"source.toby3d.me/toby3d/auth/internal/domain"
	"source.toby3d.me/toby3d/form"
)

type (
	TokenExchangeRequest struct {
		ClientID     domain.ClientID  `form:"client_id"`
		RedirectURI  domain.URL       `form:"redirect_uri"`
		GrantType    domain.GrantType `form:"grant_type"`
		Code         string           `form:"code"`
		CodeVerifier string           `form:"code_verifier"`
	}

	TokenRefreshRequest struct {
		GrantType domain.GrantType `form:"grant_type"` // refresh_token

		// The client ID that was used when the refresh token was issued.
		ClientID domain.ClientID `form:"client_id"`

		// The client may request a token with the same or fewer scopes
		// than the original access token. If omitted, is treated as
		// equal to the original scopes granted.
		Scope domain.Scopes `form:"scope"`

		// The refresh token previously offered to the client.
		RefreshToken string `form:"refresh_token"`
	}

	TokenRevocationRequest struct {
		Action domain.Action `form:"action,omitempty"`
		Token  string        `form:"token"`
	}

	TokenTicketRequest struct {
		Action domain.Action `form:"action"`
		Ticket string        `form:"ticket"`
	}

	TokenIntrospectRequest struct {
		Token string `form:"token"`
	}

	//nolint:tagliatelle // https://indieauth.net/source/#access-token-response
	TokenExchangeResponse struct {
		// The user's profile information.
		Profile *TokenProfileResponse `json:"profile,omitempty"`

		// The OAuth 2.0 Bearer Token RFC6750.
		AccessToken string `json:"access_token"`

		// The canonical user profile URL for the user this access token
		// corresponds to.
		Me string `json:"me"`

		// The refresh token, which can be used to obtain new access
		// tokens.
		RefreshToken string `json:"refresh_token"`

		// The lifetime in seconds of the access token.
		ExpiresIn int64 `json:"expires_in,omitempty"`
	}

	TokenProfileResponse struct {
		// Name the user wishes to provide to the client.
		Name string `json:"name,omitempty"`

		// URL of the user's website.
		URL string `json:"url,omitempty"`

		// A photo or image that the user wishes clients to use as a
		// profile image.
		Photo string `json:"photo,omitempty"`

		// The email address a user wishes to provide to the client.
		Email string `json:"email,omitempty"`
	}

	//nolint:tagliatelle // https://indieauth.net/source/#access-token-verification-response
	TokenIntrospectResponse struct {
		// The profile URL of the user corresponding to this token.
		Me string `json:"me"`

		// The client ID associated with this token.
		ClientID string `json:"client_id"`

		// A space-separated list of scopes associated with this token.
		Scope string `json:"scope"`

		// Integer timestamp, measured in the number of seconds since
		// January 1 1970 UTC, indicating when this token will expire.
		Exp int64 `json:"exp,omitempty"`

		// Integer timestamp, measured in the number of seconds since
		// January 1 1970 UTC, indicating when this token was originally
		// issued.
		Iat int64 `json:"iat,omitempty"`

		// Boolean indicator of whether or not the presented token is
		// currently active.
		Active bool `json:"active"`
	}

	TokenInvalidIntrospectResponse struct {
		Active bool `json:"active"`
	}

	TokenRevocationResponse struct{}
)

func (r *TokenExchangeRequest) bind(req *http.Request) error {
	indieAuthError := new(domain.Error)

	if err := form.Unmarshal([]byte(req.URL.Query().Encode()), r); err != nil {
		if errors.As(err, indieAuthError) {
			return indieAuthError
		}

		return domain.NewError(
			domain.ErrorCodeInvalidRequest,
			err.Error(),
			"https://indieauth.net/source/#request",
		)
	}

	return nil
}

func NewTokenRevocationRequest() *TokenRevocationRequest {
	return &TokenRevocationRequest{
		Action: domain.ActionRevoke,
		Token:  "",
	}
}

func (r *TokenRevocationRequest) bind(req *http.Request) error {
	indieAuthError := new(domain.Error)

	if err := req.ParseForm(); err != nil {
		return domain.NewError(
			domain.ErrorCodeInvalidRequest,
			err.Error(),
			"https://indieauth.net/source/#authorization-request",
		)
	}

	err := form.Unmarshal([]byte(req.PostForm.Encode()), r)
	if err != nil {
		if errors.As(err, indieAuthError) {
			return indieAuthError
		}

		return domain.NewError(
			domain.ErrorCodeInvalidRequest,
			err.Error(),
			"https://indieauth.net/source/#request",
		)
	}

	return nil
}

func (r *TokenTicketRequest) bind(req *http.Request) error {
	indieAuthError := new(domain.Error)

	if err := form.Unmarshal([]byte(req.URL.Query().Encode()), r); err != nil {
		if errors.As(err, indieAuthError) {
			return indieAuthError
		}

		return domain.NewError(
			domain.ErrorCodeInvalidRequest,
			err.Error(),
			"https://indieauth.net/source/#request",
		)
	}

	return nil
}

func (r *TokenIntrospectRequest) bind(req *http.Request) error {
	indieAuthError := new(domain.Error)

	if err := req.ParseForm(); err != nil {
		return domain.NewError(
			domain.ErrorCodeInvalidRequest,
			err.Error(),
			"https://indieauth.net/source/#authorization-request",
		)
	}

	if err := form.Unmarshal([]byte(req.PostForm.Encode()), r); err != nil {
		if errors.As(err, indieAuthError) {
			return indieAuthError
		}

		return domain.NewError(
			domain.ErrorCodeInvalidRequest,
			err.Error(),
			"https://indieauth.net/source/#access-token-verification-request",
		)
	}

	return nil
}
