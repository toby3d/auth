//nolint: gosec
package domain

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"hash"
	"io"
)

type (
	PKCE struct {
		Method    PKCEMethod
		Verifier  string
		Challenge string
	}

	PKCEMethod string
)

const (
	PKCEMethodMD5   PKCEMethod = "MD5"
	PKCEMethodPlain PKCEMethod = "plain"
	PKCEMethodS1    PKCEMethod = "S1"
	PKCEMethodS256  PKCEMethod = "S256"
	PKCEMethodS512  PKCEMethod = "S512"
)

func (pkce PKCE) IsValid() bool {
	h := pkce.Method.Hash()
	if h == nil { // NOTE(toby3d): PLAIN
		return pkce.Challenge == pkce.Verifier
	}

	_, _ = io.WriteString(h, pkce.Verifier)

	return pkce.Challenge == base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}

func (m PKCEMethod) Hash() hash.Hash {
	switch m {
	case PKCEMethodMD5:
		return md5.New()
	case PKCEMethodS1:
		return sha1.New()
	case PKCEMethodS256:
		return sha256.New()
	case PKCEMethodS512:
		return sha512.New()
	case PKCEMethodPlain:
		fallthrough
	default:
		return nil
	}
}
