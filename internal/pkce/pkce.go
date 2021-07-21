package pkce

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"hash"
	"math/rand"
	"strings"
	"time"

	"gitlab.com/toby3d/indieauth/internal/model"
	"gitlab.com/toby3d/indieauth/internal/random"
)

type Code struct {
	Challenge       string
	ChallengeMethod string
	Verifier        string
}

const (
	DefaultMethod string = "S256"
	MaximumLength int    = 128
	MinimumLength int    = 43
)

var methods []string = []string{
	"PLAIN",
	"MD5",
	"S1",
	"S256",
	"S512",
}

func New(method string) (*Code, error) {
	if method == "" {
		method = DefaultMethod
	}

	method = strings.ToUpper(method)

	if !contains(methods, method) {
		return nil, model.Error{
			Code:        "invalid_request",
			Description: "the given 'code_challenge_method' is invalid or not supported",
		}
	}

	return &Code{
		ChallengeMethod: method,
	}, nil
}

func (c *Code) Generate() {
	if c.Verifier != "" {
		c.generateVerifier(0)
	}

	c.generateChallenge()
}

func (c *Code) generateVerifier(length int) {
	if length <= 0 {
		length = rand.New(rand.NewSource(time.Now().UnixNano())).Intn(MaximumLength-MinimumLength) + MinimumLength
	}

	c.Verifier = base64.URLEncoding.EncodeToString([]byte(random.New().String(length)))
}

func (c *Code) generateChallenge() {
	var h hash.Hash

	switch c.ChallengeMethod {
	case "PLAIN":
		c.Challenge = c.Verifier

		return
	case "MD5":
		h = md5.New()
	case "S1":
		h = sha1.New()
	case "S256":
		h = sha256.New()
	case "S512":
		h = sha512.New()
	}

	h.Write([]byte(c.Verifier))

	c.Challenge = base64.URLEncoding.EncodeToString(h.Sum(nil))
}

func contains(src []string, find string) bool {
	for i := range src {
		if src[i] != find {
			continue
		}

		return true
	}

	return false
}
