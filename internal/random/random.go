package random

import (
	"crypto/rand"
	"fmt"
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

func Bytes(length uint8) ([]byte, error) {
	bytes := make([]byte, length)

	if _, err := rand.Read(bytes); err != nil {
		return nil, fmt.Errorf("cannot read bytes: %w", err)
	}

	return bytes, nil
}

func String(length uint8, charsets ...string) (string, error) {
	charset := strings.Join(charsets, "")
	if charset == "" {
		charset = Alphabetic
	}

	bytes := make([]byte, length)

	for i := range bytes {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", fmt.Errorf("failed to randomize bytes: %w", err)
		}

		bytes[i] = charset[n.Int64()]
	}

	return string(bytes), nil
}
