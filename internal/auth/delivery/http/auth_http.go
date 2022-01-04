package http

import (
	"net/url"
	"time"

	"github.com/fasthttp/router"
	json "github.com/goccy/go-json"
	http "github.com/valyala/fasthttp"
	"gitlab.com/toby3d/indieauth/internal/auth"
	"gitlab.com/toby3d/indieauth/internal/domain"
	"gitlab.com/toby3d/indieauth/internal/middleware"
	"gitlab.com/toby3d/indieauth/internal/pkce"
	"gitlab.com/toby3d/indieauth/web"
)

type (
	Handler struct {
		useCase auth.UseCase
	}

	AuthorizeRequest struct {
		RedirectURI         string
		ResponseType        string
		ClientID            string
		State               []byte
		Scope               string
		CodeChallenge       string
		CodeChallengeMethod string
		Me                  string
	}

	RedirectRequest struct {
		Authorize           string
		ClientID            string
		CodeChallenge       string
		CodeChallengeMethod string
		Me                  string
		RedirectURI         string
		ResponseType        string
		Scope               string
		State               []byte
	}

	ExchangeRequest struct {
		GrantType    string
		Code         string
		ClientID     string
		RedirectURI  string
		CodeVerifier string
	}

	ExchangeResponse struct {
		Me string `json:"me"`
	}
)

func NewAuthHandler(useCase auth.UseCase) *Handler {
	return &Handler{
		useCase: useCase,
	}
}

func (h *Handler) Register(r *router.Router) {
	chain := middleware.Chain{middleware.CSRFWithConfig(middleware.CSRFConfig{
		ContextKey:     "csrf",
		CookieHTTPOnly: true,
		CookieName:     "__Host-CSRF",
		CookiePath:     "/",
		CookieSameSite: http.CookieSameSiteLaxMode,
		CookieSecure:   true,
		TokenLookup:    "form:_csrf",
		Skipper: func(ctx *http.RequestCtx) bool {
			return ctx.IsPost() && ctx.PostArgs().Has("grant_type") &&
				string(ctx.PostArgs().Peek("grant_type")) == "authorization_code"
		},
	})}

	r.GET("/authorize", chain.RequestHandler(h.ClientInfo))
	r.POST("/authorize", chain.RequestHandler(h.Update))
}

func (r *AuthorizeRequest) bind(ctx *http.RequestCtx) error {
	if r.ClientID = string(ctx.QueryArgs().Peek("client_id")); r.ClientID == "" {
		return domain.Error{
			Code:        domain.ErrInvalidRequest.Code,
			Description: "'client_id' query is required",
		}
	}

	if r.ResponseType = string(ctx.QueryArgs().Peek("response_type")); r.ResponseType != "code" {
		return domain.Error{
			Code:        domain.ErrInvalidRequest.Code,
			Description: "'response_type' must be 'code'",
		}
	}

	if ctx.QueryArgs().Has("code_challenge") {
		r.CodeChallenge = string(ctx.QueryArgs().Peek("code_challenge"))
		if len(r.CodeChallenge) < 43 || len(r.CodeChallenge) > 128 {
			return domain.Error{
				Code:        domain.ErrInvalidRequest.Code,
				Description: "length of the 'code_challenge' value must be greater than 43 and less than 128 symbols",
			}
		}

		r.CodeChallengeMethod = pkce.DefaultMethod
		if ctx.PostArgs().Has("code_challenge_method") {
			r.CodeChallengeMethod = string(ctx.QueryArgs().Peek("code_challenge_method"))
		}

		if _, err := pkce.New(r.CodeChallengeMethod); err != nil {
			return domain.Error{
				Code:        domain.ErrInvalidRequest.Code,
				Description: err.Error(),
			}
		}
	}

	r.RedirectURI = string(ctx.QueryArgs().Peek("redirect_uri"))
	r.State = ctx.QueryArgs().Peek("state")
	r.Scope = string(ctx.QueryArgs().Peek("scope"))
	r.Me = string(ctx.QueryArgs().Peek("me"))

	return nil
}

func (h *Handler) ClientInfo(ctx *http.RequestCtx) {
	r := new(AuthorizeRequest)
	r.Scope = "profile"

	if err := r.bind(ctx); err != nil {
		ctx.Error(err.Error(), http.StatusBadRequest)

		return
	}

	client, err := h.useCase.Discovery(ctx, r.ClientID)
	if err != nil {
		ctx.Error(err.Error(), http.StatusBadRequest)

		return
	}

	csrf, _ := ctx.UserValue("csrf").([]byte)

	ctx.SetContentType("text/html")
	web.WritePageTemplate(ctx, &web.AuthPage{
		Client:              client,
		CodeChallenge:       r.CodeChallenge,
		CodeChallengeMethod: r.CodeChallengeMethod,
		CSRF:                csrf,
		Me:                  r.Me,
		RedirectURI:         r.RedirectURI,
		ResponseType:        r.ResponseType,
		Scope:               r.Scope,
		State:               r.State,
	})
}

func (h *Handler) Update(ctx *http.RequestCtx) {
	if ctx.PostArgs().Has("response_type") && string(ctx.PostArgs().Peek("response_type")) == "code" {
		h.Redirect(ctx)

		return
	}

	if ctx.PostArgs().Has("grant_type") && string(ctx.PostArgs().Peek("grant_type")) == "authorization_code" {
		h.Exchange(ctx)

		return
	}

	ctx.Error("please, restart your authoriztion flow", http.StatusBadRequest)
}

func (r *RedirectRequest) bind(ctx *http.RequestCtx) (err error) {
	r.RedirectURI = string(ctx.PostArgs().Peek("redirect_uri"))

	if r.ClientID = string(ctx.PostArgs().Peek("client_id")); r.ClientID == "" {
		return domain.Error{
			Code:        domain.ErrInvalidRequest.Code,
			Description: "'client_id' query is required",
		}
	}

	if r.Authorize = string(ctx.PostArgs().Peek("authorize")); r.Authorize != "allow" && r.Authorize != "deny" {
		return domain.Error{
			Code:        domain.ErrInvalidRequest.Code,
			Description: "invalid prompt action, try starting the authorization flow again",
		}
	}

	if r.ResponseType = string(ctx.PostArgs().Peek("response_type")); r.ResponseType != "code" {
		return domain.Error{
			Code:        domain.ErrInvalidRequest.Code,
			Description: "'response_type' must be 'code', try starting the authorization flow again",
		}
	}

	if ctx.PostArgs().Has("code_challenge") {
		r.CodeChallenge = string(ctx.PostArgs().Peek("code_challenge"))

		if len(r.CodeChallenge) < 43 || len(r.CodeChallenge) > 128 {
			return domain.Error{
				Code:        domain.ErrInvalidRequest.Code,
				Description: "length of the 'code_challenge' value must be greater than 43 and less than 128 symbols, try starting the authorization flow again",
			}
		}

		r.CodeChallengeMethod = pkce.DefaultMethod
		if ctx.PostArgs().Has("code_challenge_method") {
			r.CodeChallengeMethod = string(ctx.PostArgs().Peek("code_challenge_method"))
		}

		_, err := pkce.New(r.CodeChallengeMethod)
		if err != nil {
			return domain.Error{
				Code:        domain.ErrInvalidRequest.Code,
				Description: err.Error(),
			}
		}
	}

	r.State = ctx.PostArgs().Peek("state")
	r.Scope = string(ctx.PostArgs().Peek("scope"))
	r.Me = string(ctx.PostArgs().Peek("me"))

	return nil
}

func (h *Handler) Redirect(ctx *http.RequestCtx) {
	r := new(RedirectRequest)
	if err := r.bind(ctx); err != nil {
		ctx.Error(err.Error(), http.StatusBadRequest)

		return
	}

	redirectUri, err := url.Parse(r.RedirectURI)
	if err != nil {
		ctx.Error(err.Error(), http.StatusBadRequest)

		return
	}

	query := redirectUri.Query()
	query.Set("state", string(r.State))

	switch r.Authorize {
	case "allow":
		code, err := h.useCase.Approve(ctx, &domain.Login{
			CreatedAt:           time.Now().UTC().Unix(),
			ClientID:            r.ClientID,
			CodeChallenge:       r.CodeChallenge,
			CodeChallengeMethod: r.CodeChallengeMethod,
			Me:                  r.Me,
			RedirectURI:         r.RedirectURI,
			Scope:               r.Scope,
		})
		if err != nil {
			query.Set("error", domain.ErrServerError.Code)
			query.Set("error_description", err.Error())

			redirectUri.RawQuery = query.Encode()

			ctx.Redirect(redirectUri.String(), http.StatusFound)

			return
		}

		query.Set("code", code)
	case "deny":
		query.Set("error", domain.ErrAccessDenied.Code)
	}

	redirectUri.RawQuery = query.Encode()

	ctx.Redirect(redirectUri.String(), http.StatusFound)
}

func (r *ExchangeRequest) bind(ctx *http.RequestCtx) (err error) {
	if r.GrantType = string(ctx.PostArgs().Peek("grant_type")); r.GrantType != "authorization_code" {
		return domain.Error{
			Code:        domain.ErrInvalidRequest.Code,
			Description: "'grant_type' must be 'authorization_code'",
		}
	}

	if r.RedirectURI = string(ctx.PostArgs().Peek("redirect_uri")); r.RedirectURI == "" {
		return domain.Error{
			Code:        domain.ErrInvalidRequest.Code,
			Description: "'redirect_uri' query is required",
		}
	}

	if r.ClientID = string(ctx.PostArgs().Peek("client_id")); r.ClientID == "" {
		return domain.Error{
			Code:        domain.ErrInvalidRequest.Code,
			Description: "'client_id' query is required",
		}
	}

	if r.Code = string(ctx.PostArgs().Peek("code")); r.Code == "" {
		return domain.Error{
			Code:        domain.ErrInvalidRequest.Code,
			Description: "'code' query is required",
		}
	}

	r.CodeVerifier = string(ctx.PostArgs().Peek("code_verifier"))

	return nil
}

func (h *Handler) Exchange(ctx *http.RequestCtx) {
	encoder := json.NewEncoder(ctx)
	req := new(ExchangeRequest)

	if err := req.bind(ctx); err != nil {
		ctx.Error(err.Error(), http.StatusBadRequest)

		return
	}

	me, err := h.useCase.Exchange(ctx, &domain.ExchangeRequest{
		ClientID:     req.ClientID,
		Code:         req.Code,
		CodeVerifier: req.CodeVerifier,
		RedirectURI:  req.RedirectURI,
	})
	if err != nil {
		ctx.Error(err.Error(), http.StatusBadRequest)

		return
	}

	if me == "" {
		ctx.Error(domain.ErrUnauthorizedClient.Error(), http.StatusUnauthorized)

		return
	}

	ctx.SetContentType("application/json")
	_ = encoder.Encode(&ExchangeResponse{
		Me: me,
	})
}
