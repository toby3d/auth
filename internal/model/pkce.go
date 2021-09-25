package model

import (
	"bytes"
	"hash"
)

type PKCE struct {
	Challenge string
	Method    hash.Hash
	Verifier  string
}

func (pkce PKCE) IsValid() bool {
	return bytes.Equal([]byte(pkce.Challenge), pkce.Method.Sum([]byte(pkce.Verifier)))
}
