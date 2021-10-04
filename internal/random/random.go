package random

import (
	"math/rand"
	"strings"
	"time"
)

type Random struct{}

const (
	Uppercase    = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	Lowercase    = "abcdefghijklmnopqrstuvwxyz"
	Alphabetic   = Uppercase + Lowercase
	Numeric      = "0123456789"
	Alphanumeric = Alphabetic + Numeric
	Symbols      = "`" + `~!@#$%^&*()-_+={}[]|\;:"<>,./?`
	Hex          = Numeric + "abcdef"
)

func New() *Random {
	rand.Seed(time.Now().UnixNano())

	return new(Random)
}

func (r *Random) String(length int, charsets ...string) string {
	charset := strings.Join(charsets, "")

	if charset == "" {
		charset = Alphabetic
	}

	b := make([]byte, length)

	for i := range b {
		//nolint: gosec
		b[i] = charset[rand.Int()%len(charset)]
	}

	return string(b)
}
