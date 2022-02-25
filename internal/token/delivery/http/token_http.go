package http

import (
	"errors"
	"path"

	"github.com/fasthttp/router"
	json "github.com/goccy/go-json"
	"github.com/lestrrat-go/jwx/jwa"
	http "github.com/valyala/fasthttp"

	"source.toby3d.me/toby3d/form"
	"source.toby3d.me/toby3d/middleware"
	"source.toby3d.me/website/indieauth/internal/common"
	"source.toby3d.me/website/indieauth/internal/domain"
	"source.toby3d.me/website/indieauth/internal/ticket"
	"source.toby3d.me/website/indieauth/internal/token"
)

type (
	TokenExchangeRequest struct {
		ClientID     *domain.ClientID `form:"client_id"`
		RedirectURI  *domain.URL      `form:"redirect_uri"`
		GrantType    domain.GrantType `form:"grant_type"`
		Code         string           `form:"code"`
		CodeVerifier string           `form:"code_verifier"`
	}

	TokenRefreshRequest struct {
		GrantType domain.GrantType `form:"grant_type"` // refresh_token

		// The refresh token previously offered to the client.
		RefreshToken string `form:"refresh_token"`

		// The client ID that was used when the refresh token was issued.
		ClientID *domain.ClientID `form:"client_id"`

		// The client may request a token with the same or fewer scopes
		// than the original access token. If omitted, is treated as
		// equal to the original scopes granted.
		Scope domain.Scopes `form:"scope,omitempty"`
	}

	TokenRevocationRequest struct {
		Action domain.Action `form:"action"`
		Token  string        `form:"token"`
	}

	TokenTicketRequest struct {
		Action domain.Action `form:"action"`
		Ticket string        `form:"ticket"`
	}

	TokenIntrospectRequest struct {
		Token string `form:"token"`
	}

	//nolint: tagliatelle // https://indieauth.net/source/#access-token-response
	TokenExchangeResponse struct {
		// The OAuth 2.0 Bearer Token RFC6750.
		AccessToken string `json:"access_token"`

		// The canonical user profile URL for the user this access token
		// corresponds to.
		Me string `json:"me"`

		// The user's profile information.
		Profile *TokenProfileResponse `json:"profile,omitempty"`

		// The lifetime in seconds of the access token.
		ExpiresIn int64 `json:"expires_in,omitempty"`

		// The refresh token, which can be used to obtain new access
		// tokens.
		RefreshToken string `json:"refresh_token"`
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

	//nolint: tagliatelle // https://indieauth.net/source/#access-token-verification-response
	TokenIntrospectResponse struct {
		// Boolean indicator of whether or not the presented token is
		// currently active.
		Active bool `json:"active"`

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
	}

	TokenInvalidIntrospectResponse struct {
		Active bool `json:"active"`
	}

	TokenRevocationResponse struct{}

	RequestHandler struct {
		config  *domain.Config
		tokens  token.UseCase
		tickets ticket.UseCase
	}
)

func NewRequestHandler(tokens token.UseCase, tickets ticket.UseCase, config *domain.Config) *RequestHandler {
	return &RequestHandler{
		config:  config,
		tokens:  tokens,
		tickets: tickets,
	}
}

func (h *RequestHandler) Register(r *router.Router) {
	chain := middleware.Chain{
		middleware.JWTWithConfig(middleware.JWTConfig{
			AuthScheme:    "Bearer",
			ContextKey:    "token",
			SigningKey:    []byte(h.config.JWT.Secret),
			SigningMethod: jwa.SignatureAlgorithm(h.config.JWT.Algorithm),
			Skipper: func(ctx *http.RequestCtx) bool {
				matched, _ := path.Match("/token*", string(ctx.Path()))

				return matched
			},
			SuccessHandler: nil,
			TokenLookup: middleware.SourceHeader + ":" + http.HeaderAuthorization +
				"," + middleware.SourceParam + ":" + "token",
		}),
		middleware.LogFmt(),
	}

	r.POST("/token", chain.RequestHandler(h.handleAction))
	r.POST("/introspect", chain.RequestHandler(h.handleIntrospect))
	r.POST("/revocation", chain.RequestHandler(h.handleRevokation))
}

func (h *RequestHandler) handleIntrospect(ctx *http.RequestCtx) {
	ctx.SetContentType(common.MIMEApplicationJSONCharsetUTF8)
	ctx.SetStatusCode(http.StatusOK)

	encoder := json.NewEncoder(ctx)

	req := new(TokenIntrospectRequest)
	if err := req.bind(ctx); err != nil {
		ctx.SetStatusCode(http.StatusBadRequest)

		_ = encoder.Encode(err)

		return
	}

	tkn, _, err := h.tokens.Verify(ctx, req.Token)
	if err != nil || tkn == nil {
		// WARN(toby3d): If the token is not valid, the endpoint still
		// MUST return a 200 Response.
		_ = encoder.Encode(&TokenInvalidIntrospectResponse{
			Active: false,
		})

		return
	}

	_ = encoder.Encode(&TokenIntrospectResponse{
		Active:   true,
		ClientID: tkn.ClientID.String(),
		Exp:      tkn.Expiry.Unix(),
		Iat:      tkn.CreatedAt.Unix(),
		Me:       tkn.Me.String(),
		Scope:    tkn.Scope.String(),
	})
}

func (h *RequestHandler) handleAction(ctx *http.RequestCtx) {
	ctx.SetContentType(common.MIMEApplicationJSONCharsetUTF8)

	encoder := json.NewEncoder(ctx)

	switch {
	case ctx.PostArgs().Has("grant_type"):
		h.handleExchange(ctx)
	case ctx.PostArgs().Has("action"):
		action, err := domain.ParseAction(string(ctx.PostArgs().Peek("action")))
		if err != nil {
			ctx.SetStatusCode(http.StatusBadRequest)

			_ = encoder.Encode(domain.NewError(
				domain.ErrorCodeInvalidRequest,
				err.Error(),
				"",
			))

			return
		}

		switch action {
		case domain.ActionRevoke:
			h.handleRevokation(ctx)
		case domain.ActionTicket:
			h.handleTicket(ctx)
		}
	}
}

func (h *RequestHandler) handleExchange(ctx *http.RequestCtx) {
	ctx.SetContentType(common.MIMEApplicationJSONCharsetUTF8)

	encoder := json.NewEncoder(ctx)

	req := new(TokenExchangeRequest)
	if err := req.bind(ctx); err != nil {
		ctx.SetStatusCode(http.StatusBadRequest)

		_ = encoder.Encode(err)

		return
	}

	token, profile, err := h.tokens.Exchange(ctx, token.ExchangeOptions{
		ClientID:     req.ClientID,
		RedirectURI:  req.RedirectURI,
		Code:         req.Code,
		CodeVerifier: req.CodeVerifier,
	})
	if err != nil {
		ctx.SetStatusCode(http.StatusBadRequest)

		_ = encoder.Encode(domain.NewError(
			domain.ErrorCodeInvalidRequest,
			err.Error(),
			"https://indieauth.net/source/#request",
		))

		return
	}

	resp := &TokenExchangeResponse{
		AccessToken:  token.AccessToken,
		ExpiresIn:    token.Expiry.Unix(),
		Me:           token.Me.String(),
		Profile:      nil,
		RefreshToken: "", // TODO(toby3d)
	}

	if profile == nil {
		_ = encoder.Encode(resp)

		return
	}

	resp.Profile = &TokenProfileResponse{
		Name:  profile.GetName(),
		URL:   "",
		Photo: "",
		Email: "",
	}

	if url := profile.GetURL(); url != nil {
		resp.Profile.URL = url.String()
	}

	if photo := profile.GetPhoto(); photo != nil {
		resp.Profile.Photo = photo.String()
	}

	if email := profile.GetEmail(); email != nil {
		resp.Profile.Email = email.String()
	}

	_ = encoder.Encode(resp)
}

func (h *RequestHandler) handleRevokation(ctx *http.RequestCtx) {
	ctx.SetContentType(common.MIMEApplicationJSONCharsetUTF8)
	ctx.SetStatusCode(http.StatusOK)

	encoder := json.NewEncoder(ctx)

	req := new(TokenRevocationRequest)
	if err := req.bind(ctx); err != nil {
		ctx.SetStatusCode(http.StatusBadRequest)

		_ = encoder.Encode(err)

		return
	}

	if err := h.tokens.Revoke(ctx, req.Token); err != nil {
		ctx.SetStatusCode(http.StatusBadRequest)

		_ = encoder.Encode(domain.NewError(
			domain.ErrorCodeInvalidRequest,
			err.Error(),
			"",
		))

		return
	}

	_ = encoder.Encode(&TokenRevocationResponse{})
}

func (h *RequestHandler) handleTicket(ctx *http.RequestCtx) {
	ctx.SetContentType(common.MIMEApplicationJSONCharsetUTF8)
	ctx.SetStatusCode(http.StatusOK)

	encoder := json.NewEncoder(ctx)

	req := new(TokenTicketRequest)
	if err := req.bind(ctx); err != nil {
		ctx.SetStatusCode(http.StatusBadRequest)

		_ = encoder.Encode(err)

		return
	}

	tkn, err := h.tickets.Exchange(ctx, req.Ticket)
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)

		_ = encoder.Encode(domain.NewError(
			domain.ErrorCodeInvalidRequest,
			err.Error(),
			"https://indieauth.net/source/#request",
		))

		return
	}

	_ = encoder.Encode(&TokenExchangeResponse{
		AccessToken:  tkn.AccessToken,
		Me:           tkn.Me.String(),
		Profile:      nil,
		ExpiresIn:    tkn.Expiry.Unix(),
		RefreshToken: "", // TODO(toby3d)
	})
}

func (r *TokenExchangeRequest) bind(ctx *http.RequestCtx) error {
	indieAuthError := new(domain.Error)
	if err := form.Unmarshal(ctx.PostArgs(), r); err != nil {
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

func (r *TokenRevocationRequest) bind(ctx *http.RequestCtx) error {
	indieAuthError := new(domain.Error)
	if err := form.Unmarshal(ctx.PostArgs(), r); err != nil {
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

func (r *TokenTicketRequest) bind(ctx *http.RequestCtx) error {
	indieAuthError := new(domain.Error)
	if err := form.Unmarshal(ctx.PostArgs(), r); err != nil {
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

func (r *TokenIntrospectRequest) bind(ctx *http.RequestCtx) error {
	indieAuthError := new(domain.Error)
	if err := form.Unmarshal(ctx.PostArgs(), r); err != nil {
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
