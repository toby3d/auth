package domain_test

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	http "github.com/valyala/fasthttp"

	"source.toby3d.me/toby3d/auth/internal/domain"
)

func TestNewToken(t *testing.T) {
	t.Parallel()

	expResult := domain.TestToken(t)
	opts := domain.NewTokenOptions{
		Algorithm:   "",
		NonceLength: 0,
		Issuer:      expResult.ClientID,
		Expiration:  1 * time.Hour,
		Scope:       expResult.Scope,
		Subject:     expResult.Me,
		Secret:      []byte("hackme"),
	}

	result, err := domain.NewToken(opts)
	if err != nil {
		t.Fatalf("%+v", err)
	}

	if fmt.Sprint(result.ClientID) != fmt.Sprint(expResult.ClientID) ||
		fmt.Sprint(result.Me) != fmt.Sprint(expResult.Me) ||
		fmt.Sprint(result.Scope) != fmt.Sprint(expResult.Scope) {
		t.Errorf("NewToken(%+v) = %+v, want %+v", opts, result, expResult)
	}
}

func TestToken_SetAuthHeader(t *testing.T) {
	t.Parallel()

	token := domain.TestToken(t)
	expResult := []byte("Bearer " + token.AccessToken)

	req := http.AcquireRequest()
	defer http.ReleaseRequest(req)
	token.SetAuthHeader(req)

	result := req.Header.Peek(http.HeaderAuthorization)
	if result == nil || !bytes.Equal(result, expResult) {
		t.Errorf("SetAuthHeader(%+v) = %s, want %s", req, result, expResult)
	}
}

func TestToken_String(t *testing.T) {
	t.Parallel()

	token := domain.TestToken(t)
	expResult := "Bearer " + token.AccessToken

	if result := token.String(); result != expResult {
		t.Errorf("String() = %s, want %s", result, expResult)
	}
}
