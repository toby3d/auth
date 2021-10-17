package domain

import (
	"strings"
	"testing"
	"time"

	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwt"
	"github.com/stretchr/testify/require"

	"source.toby3d.me/website/oauth/internal/random"
)

type Token struct {
	Expiry      time.Time
	Scopes      []string
	AccessToken string
	TokenType   string
	ClientID    string
	Me          string
}

func NewToken() *Token {
	t := new(Token)
	t.Expiry = time.Time{}

	return t
}

func TestToken(tb testing.TB) *Token {
	tb.Helper()

	require := require.New(tb)

	nonce, err := random.String(50)
	require.NoError(err)

	client := TestClient(tb)
	profile := TestProfile(tb)
	now := time.Now().UTC().Round(time.Second)
	scopes := []string{"create", "update", "delete"}
	t := jwt.New()

	// required
	t.Set(jwt.IssuerKey, client.ID)    // NOTE(toby3d): client_id
	t.Set(jwt.SubjectKey, profile.URL) // NOTE(toby3d): me
	// TODO(toby3d): t.Set(jwt.AudienceKey, nil)
	t.Set(jwt.ExpirationKey, now.Add(1*time.Hour))
	t.Set(jwt.NotBeforeKey, now.Add(-1*time.Hour))
	t.Set(jwt.IssuedAtKey, now.Add(-1*time.Hour))
	// TODO(toby3d): t.Set(jwt.JwtIDKey, nil)

	// optional
	t.Set("scope", strings.Join(scopes, " "))
	t.Set("nonce", nonce)

	accessToken, err := jwt.Sign(t, jwa.HS256, []byte("hackme"))
	require.NoError(err)

	return &Token{
		AccessToken: string(accessToken),
		ClientID:    t.Issuer(),
		Expiry:      t.Expiration(),
		Me:          t.Subject(),
		Scopes:      scopes,
		TokenType:   "Bearer",
	}
}
