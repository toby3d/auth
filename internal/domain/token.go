package domain

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwt"

	"source.toby3d.me/toby3d/auth/internal/common"
	"source.toby3d.me/toby3d/auth/internal/random"
)

type (
	// Token describes the data of the token used by the clients.
	Token struct {
		CreatedAt    time.Time
		Expiry       time.Time
		ClientID     ClientID
		Me           Me
		Scope        Scopes
		AccessToken  string
		RefreshToken string
	}

	// NewTokenOptions contains options for NewToken function.
	NewTokenOptions struct {
		Expiration  time.Duration
		Issuer      ClientID
		Subject     Me
		Scope       Scopes
		Secret      []byte
		Algorithm   string
		NonceLength uint8
	}
)

// DefaultNewTokenOptions describes the default settings for NewToken.
//
//nolint:gochecknoglobals,gomnd
var DefaultNewTokenOptions = NewTokenOptions{
	Expiration:  0,
	Scope:       nil,
	Issuer:      ClientID{},
	Subject:     Me{},
	Secret:      nil,
	Algorithm:   "HS256",
	NonceLength: 32,
}

// NewToken create a new token by provided options.
//
//nolint:cyclop
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

	for key, val := range map[string]any{
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

	if opts.Issuer.clientID != nil {
		if err = tkn.Set(jwt.IssuerKey, opts.Issuer.String()); err != nil {
			return nil, fmt.Errorf("failed to set JWT token field: %w", err)
		}
	}

	if opts.Expiration != 0 {
		if err = tkn.Set(jwt.ExpirationKey, now.Add(opts.Expiration)); err != nil {
			return nil, fmt.Errorf("failed to set JWT token field: %w", err)
		}
	}

	accessToken, err := jwt.Sign(tkn, jwt.WithKey(jwa.SignatureAlgorithm(opts.Algorithm), opts.Secret))
	if err != nil {
		return nil, fmt.Errorf("cannot sign a new access token: %w", err)
	}

	return &Token{
		AccessToken:  string(accessToken),
		ClientID:     opts.Issuer,
		CreatedAt:    now,
		Expiry:       now.Add(opts.Expiration),
		Me:           opts.Subject,
		RefreshToken: "", // TODO(toby3d)
		Scope:        opts.Scope,
	}, nil
}

// TestToken returns valid random generated token for tests.
//
//nolint:gomnd // testing domain can contains non-standart values
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
		ScopeProfile,
		ScopeEmail,
	}

	for key, val := range map[string]any{
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

	accessToken, err := jwt.Sign(tkn, jwt.WithKey(jwa.HS256, []byte("hackme")))
	if err != nil {
		tb.Fatal(err)
	}

	return &Token{
		CreatedAt:    now.Add(-1 * time.Hour),
		Expiry:       now.Add(1 * time.Hour),
		ClientID:     *cid,
		Me:           *me,
		Scope:        scope,
		AccessToken:  string(accessToken),
		RefreshToken: "", // TODO(toby3d)
	}
}

// SetAuthHeader writes an Access Token to the request header.
func (t Token) SetAuthHeader(r *http.Request) {
	if t.AccessToken == "" {
		return
	}

	r.Header.Set(common.HeaderAuthorization, t.String())
}

// String returns string representation of token.
func (t Token) String() string {
	if t.AccessToken == "" {
		return ""
	}

	return "Bearer " + t.AccessToken
}
