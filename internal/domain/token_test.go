package domain_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"source.toby3d.me/toby3d/auth/internal/common"
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

	in := domain.TestToken(t)
	req, _ := http.NewRequest(http.MethodGet, "https://example.com/", nil)
	in.SetAuthHeader(req)

	exp := "Bearer " + in.AccessToken
	if out := req.Header.Get(common.HeaderAuthorization); out != exp {
		t.Errorf("SetAuthHeader(%+v) = %s, want %s", req, out, exp)
	}
}

func TestToken_String(t *testing.T) {
	t.Parallel()

	token := domain.TestToken(t)
	exp := "Bearer " + token.AccessToken

	if out := token.String(); out != exp {
		t.Errorf("String() = %s, want %s", out, exp)
	}
}
