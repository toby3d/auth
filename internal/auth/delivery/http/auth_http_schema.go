package http

import (
	"errors"
	"net/http"
	"strings"

	"source.toby3d.me/toby3d/auth/internal/domain"
	"source.toby3d.me/toby3d/form"
)

type (
	AuthAuthorizationRequest struct {
		// Indicates to the authorization server that an authorization
		// code should be returned as the response.
		ResponseType domain.ResponseType `form:"response_type"` // code

		// The client URL.
		ClientID domain.ClientID `form:"client_id"`

		// The redirect URL indicating where the user should be
		// redirected to after approving the request.
		RedirectURI domain.URL `form:"redirect_uri"`

		// The URL that the user entered.
		Me domain.Me `form:"me"`

		// The hashing method used to calculate the code challenge.
		CodeChallengeMethod *domain.CodeChallengeMethod `form:"code_challenge_method,omitempty"`

		// A space-separated list of scopes the client is requesting,
		// e.g. "profile", or "profile create". If the client omits this
		// value, the authorization server MUST NOT issue an access
		// token for this authorization code. Only the user's profile
		// URL may be returned without any scope requested.
		Scope domain.Scopes `form:"scope,omitempty"`

		// A parameter set by the client which will be included when the
		// user is redirected back to the client. This is used to
		// prevent CSRF attacks. The authorization server MUST return
		// the unmodified state value back to the client.
		State string `form:"state"`

		// The code challenge as previously described.
		CodeChallenge string `form:"code_challenge,omitempty"`
	}

	AuthVerifyRequest struct {
		ClientID            domain.ClientID             `form:"client_id"`
		Me                  domain.Me                   `form:"me"`
		RedirectURI         domain.URL                  `form:"redirect_uri"`
		ResponseType        domain.ResponseType         `form:"response_type"`
		CodeChallengeMethod *domain.CodeChallengeMethod `form:"code_challenge_method,omitempty"`
		Scope               domain.Scopes               `form:"scope[],omitempty"`
		Authorize           string                      `form:"authorize"`
		CodeChallenge       string                      `form:"code_challenge,omitempty"`
		State               string                      `form:"state"`
		Provider            string                      `form:"provider"`
	}

	AuthExchangeRequest struct {
		GrantType domain.GrantType `form:"grant_type"` // authorization_code

		// The client's URL, which MUST match the client_id used in the
		// authentication request.
		ClientID domain.ClientID `form:"client_id"`

		// The client's redirect URL, which MUST match the initial
		// authentication request.
		RedirectURI domain.URL `form:"redirect_uri"`

		// The authorization code received from the authorization
		// endpoint in the redirect.
		Code string `form:"code"`

		// The original plaintext random string generated before
		// starting the authorization request.
		CodeVerifier string `form:"code_verifier"`
	}

	AuthExchangeResponse struct {
		Me      domain.Me            `json:"me"`
		Profile *AuthProfileResponse `json:"profile,omitempty"`
	}

	AuthProfileResponse struct {
		Email *domain.Email `json:"email,omitempty"`
		Photo *domain.URL   `json:"photo,omitempty"`
		URL   *domain.URL   `json:"url,omitempty"`
		Name  string        `json:"name,omitempty"`
	}
)

func NewAuthAuthorizationRequest() *AuthAuthorizationRequest {
	return &AuthAuthorizationRequest{
		ClientID:            domain.ClientID{},
		CodeChallenge:       "",
		CodeChallengeMethod: &domain.CodeChallengeMethodUnd,
		Me:                  domain.Me{},
		RedirectURI:         domain.URL{},
		ResponseType:        domain.ResponseTypeUnd,
		Scope:               make(domain.Scopes, 0),
		State:               "",
	}
}

//nolint:cyclop
func (r *AuthAuthorizationRequest) bind(req *http.Request) error {
	indieAuthError := new(domain.Error)

	if err := form.Unmarshal([]byte(req.URL.Query().Encode()), r); err != nil {
		if errors.As(err, indieAuthError) {
			return indieAuthError
		}

		return domain.NewError(
			domain.ErrorCodeInvalidRequest,
			err.Error(),
			"https://indieauth.net/source/#authorization-request",
		)
	}

	if r.ResponseType == domain.ResponseTypeID {
		r.ResponseType = domain.ResponseTypeCode
	}

	return nil
}

func NewAuthVerifyRequest() *AuthVerifyRequest {
	return &AuthVerifyRequest{
		Authorize:           "",
		ClientID:            domain.ClientID{},
		CodeChallenge:       "",
		CodeChallengeMethod: &domain.CodeChallengeMethodUnd,
		Me:                  domain.Me{},
		Provider:            "",
		RedirectURI:         domain.URL{},
		ResponseType:        domain.ResponseTypeUnd,
		Scope:               make(domain.Scopes, 0),
		State:               "",
	}
}

//nolint:funlen,cyclop
func (r *AuthVerifyRequest) bind(req *http.Request) error {
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
			"https://indieauth.net/source/#authorization-request",
		)
	}

	// NOTE(toby3d): backwards-compatible support.
	// See: https://aaronparecki.com/2020/12/03/1/indieauth-2020#response-type
	if r.ResponseType == domain.ResponseTypeID {
		r.ResponseType = domain.ResponseTypeCode
	}

	r.Provider = strings.ToLower(r.Provider)

	if !strings.EqualFold(r.Authorize, "allow") && !strings.EqualFold(r.Authorize, "deny") {
		return domain.NewError(
			domain.ErrorCodeInvalidRequest,
			"cannot validate verification request",
			"https://indieauth.net/source/#authorization-request",
		)
	}

	return nil
}

func (r *AuthExchangeRequest) bind(req *http.Request) error {
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
			"cannot validate verification request",
			"https://indieauth.net/source/#redeeming-the-authorization-code",
		)
	}

	return nil
}
