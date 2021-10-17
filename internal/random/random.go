package random

import (
	"crypto/rand"
	"math/big"
	"strings"
)

const (
	Uppercase    = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	Lowercase    = "abcdefghijklmnopqrstuvwxyz"
	Alphabetic   = Uppercase + Lowercase
	Numeric      = "0123456789"
	Alphanumeric = Alphabetic + Numeric
	Symbols      = "`" + `~!@#$%^&*()-_+={}[]|\;:"<>,./?`
	Hex          = Numeric + "abcdef"
)

func Bytes(length int) ([]byte, error) {
	b := make([]byte, length)

	if _, err := rand.Read(b); err != nil {
		return nil, err
	}

	return b, nil
}

func String(length int, charsets ...string) (string, error) {
	charset := strings.Join(charsets, "")

	if charset == "" {
		charset = Alphabetic
	}

	b := make([]byte, length)

	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}

		b[i] = charset[n.Int64()]
	}

	return string(b), nil
}
