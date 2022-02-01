package domain

import (
	"fmt"
	"testing"
	"time"

	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwt"
	http "github.com/valyala/fasthttp"

	"source.toby3d.me/website/indieauth/internal/random"
)

type (
	// Token describes the data of the token used by the clients.
	Token struct {
		Scope       Scopes
		ClientID    *ClientID
		Me          *Me
		AccessToken string
	}

	// NewTokenOptions contains options for NewToken function.
	NewTokenOptions struct {
		Expiration  time.Duration
		Scope       Scopes
		Issuer      *ClientID
		Subject     *Me
		Secret      []byte
		Algorithm   string
		NonceLength int
	}
)

// DefaultNewTokenOptions describes the default settings for NewToken.
//nolint: gochecknoglobals, gomnd
var DefaultNewTokenOptions = NewTokenOptions{
	Algorithm:   "HS256",
	Expiration:  0,
	Issuer:      nil,
	NonceLength: 32,
	Scope:       nil,
	Secret:      nil,
	Subject:     nil,
}

// NewToken create a new token by provided options.
//nolint: cyclop
func NewToken(opts NewTokenOptions) (*Token, error) {
	if opts.NonceLength == 0 {
		opts.NonceLength = DefaultNewTokenOptions.NonceLength
	}

	if opts.Algorithm == "" {
		opts.Algorithm = DefaultNewTokenOptions.Algorithm
	}

	now := time.Now().UTC().Round(time.Second)

	nonce, err := random.String(opts.NonceLength)
	if err != nil {
		return nil, fmt.Errorf("cannot generate nonce: %w", err)
	}

	tkn := jwt.New()

	for key, val := range map[string]interface{}{
		"nonce":          nonce,
		"scope":          opts.Scope,
		jwt.IssuedAtKey:  now,
		jwt.NotBeforeKey: now,
		jwt.SubjectKey:   opts.Subject.String(),
	} {
		if err = tkn.Set(key, val); err != nil {
			return nil, fmt.Errorf("failed to set JWT token field: %w", err)
		}
	}

	if opts.Issuer != nil {
		if err = tkn.Set(jwt.IssuerKey, opts.Issuer.String()); err != nil {
			return nil, fmt.Errorf("failed to set JWT token field: %w", err)
		}
	}

	if opts.Expiration != 0 {
		if err = tkn.Set(jwt.ExpirationKey, now.Add(opts.Expiration)); err != nil {
			return nil, fmt.Errorf("failed to set JWT token field: %w", err)
		}
	}

	accessToken, err := jwt.Sign(tkn, jwa.SignatureAlgorithm(opts.Algorithm), opts.Secret)
	if err != nil {
		return nil, fmt.Errorf("cannot sign a new access token: %w", err)
	}

	return &Token{
		AccessToken: string(accessToken),
		ClientID:    opts.Issuer,
		Me:          opts.Subject,
		Scope:       opts.Scope,
	}, nil
}

// TestToken returns valid random generated token for tests.
//nolint: gomnd // testing domain can contains non-standart values
func TestToken(tb testing.TB) *Token {
	tb.Helper()

	nonce, err := random.String(22)
	if err != nil {
		tb.Fatal(err)
	}

	tkn := jwt.New()
	cid := TestClientID(tb)
	me := TestMe(tb, "https://user.example.net/")
	now := time.Now().UTC().Round(time.Second)
	scope := Scopes{
		ScopeCreate,
		ScopeDelete,
		ScopeUpdate,
	}

	for key, val := range map[string]interface{}{
		// NOTE(toby3d): required
		jwt.IssuerKey:     cid.String(),
		jwt.SubjectKey:    me.String(),
		jwt.ExpirationKey: now.Add(1 * time.Hour),
		jwt.NotBeforeKey:  now.Add(-1 * time.Hour),
		jwt.IssuedAtKey:   now.Add(-1 * time.Hour),
		// TODO(toby3d): jwt.AudienceKey
		// TODO(toby3d): jwt.JwtIDKey
		// NOTE(toby3d): optional
		"scope": scope,
		"nonce": nonce,
	} {
		_ = tkn.Set(key, val)
	}

	accessToken, err := jwt.Sign(tkn, jwa.HS256, []byte("hackme"))
	if err != nil {
		tb.Fatal(err)
	}

	return &Token{
		ClientID:    cid,
		Me:          me,
		Scope:       scope,
		AccessToken: string(accessToken),
	}
}

// SetAuthHeader writes an Access Token to the request header.
func (t Token) SetAuthHeader(r *http.Request) {
	if t.AccessToken == "" {
		return
	}

	r.Header.Set(http.HeaderAuthorization, t.String())
}

// String returns string representation of token.
func (t Token) String() string {
	if t.AccessToken == "" {
		return ""
	}

	return "Bearer " + t.AccessToken
}
