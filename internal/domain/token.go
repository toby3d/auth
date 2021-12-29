package domain

import (
	"fmt"
	"testing"
	"time"

	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwt"
	"github.com/stretchr/testify/require"
	http "github.com/valyala/fasthttp"

	"source.toby3d.me/website/oauth/internal/random"
)

type (
	// Token describes the data of the token used by the clients.
	Token struct {
		AccessToken string
		ClientID    *ClientID
		Me          *Me
		Scope       Scopes
	}

	NewTokenOptions struct {
		Algorithm   string
		Expiration  time.Duration
		Issuer      *ClientID
		NonceLength int
		Scope       Scopes
		Secret      interface{}
		Subject     *Me
	}
)

var DefaultNewTokenOptions = NewTokenOptions{
	NonceLength: 32,
	Algorithm:   "HS256",
}

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

	t := jwt.New()
	t.Set(jwt.IssuerKey, opts.Issuer.String())
	t.Set(jwt.SubjectKey, opts.Subject.String())
	t.Set(jwt.NotBeforeKey, now)
	t.Set(jwt.IssuedAtKey, now)
	t.Set("scope", opts.Scope)
	t.Set("nonce", nonce)

	if opts.Expiration != 0 {
		t.Set(jwt.ExpirationKey, now.Add(opts.Expiration))
	}

	accessToken, err := jwt.Sign(t, jwa.SignatureAlgorithm(opts.Algorithm), opts.Secret)
	if err != nil {
		return nil, fmt.Errorf("cannot sign a new access token: %w", err)
	}

	return &Token{
		AccessToken: string(accessToken),
		ClientID:    opts.Issuer,
		Me:          opts.Subject,
		Scope:       opts.Scope,
	}, err
}

// TestToken returns a valid Token with the generated test data filled in.
func TestToken(tb testing.TB) *Token {
	tb.Helper()

	nonce, err := random.String(42)
	require.NoError(tb, err)

	t := jwt.New()
	cid := TestClientID(tb)
	me := TestMe(tb)
	now := time.Now().UTC().Round(time.Second)
	scope := []Scope{
		ScopeCreate,
		ScopeUpdate,
		ScopeDelete,
	}

	// NOTE(toby3d): required
	t.Set(jwt.IssuerKey, cid.String())
	t.Set(jwt.SubjectKey, me.me.String())
	// TODO(toby3d): t.Set(jwt.AudienceKey, nil)
	t.Set(jwt.ExpirationKey, now.Add(1*time.Hour))
	t.Set(jwt.NotBeforeKey, now.Add(-1*time.Hour))
	t.Set(jwt.IssuedAtKey, now.Add(-1*time.Hour))
	// TODO(toby3d): t.Set(jwt.JwtIDKey, nil)

	// optional
	t.Set("scope", scope)
	t.Set("nonce", nonce)

	accessToken, err := jwt.Sign(t, jwa.HS256, []byte("hackme"))
	require.NoError(tb, err)

	return &Token{
		ClientID:    cid,
		Me:          me,
		Scope:       scope,
		AccessToken: string(accessToken),
	}
}

// SetAuthHeader writes an Access Token to the request header.
func (t *Token) SetAuthHeader(r *http.Request) {
	if t.AccessToken == "" {
		return
	}

	r.Header.Set(http.HeaderAuthorization, t.String())
}

func (t *Token) String() string {
	if t.AccessToken == "" {
		return ""
	}

	return "Bearer " + string(t.AccessToken)
}
