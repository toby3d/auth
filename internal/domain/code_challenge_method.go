package domain

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"errors"
	"fmt"
	"hash"
	"strings"
)

// NOTE(toby3d): Encapsulate enums in structs for extra compile-time safety:
// https://threedots.tech/post/safer-enums-in-go/#struct-based-enums
type CodeChallengeMethod struct {
	hash hash.Hash
	slug string
}

//nolint: gochecknoglobals // NOTE(toby3d): structs cannot be constants
var (
	CodeChallengeMethodUndefined = CodeChallengeMethod{
		slug: "",
		hash: nil,
	}

	CodeChallengeMethodPLAIN = CodeChallengeMethod{
		slug: "PLAIN",
		hash: nil,
	}

	CodeChallengeMethodMD5 = CodeChallengeMethod{
		slug: "MD5",
		hash: md5.New(),
	}

	CodeChallengeMethodS1 = CodeChallengeMethod{
		slug: "S1",
		hash: sha1.New(),
	}

	CodeChallengeMethodS256 = CodeChallengeMethod{
		slug: "S256",
		hash: sha256.New(),
	}

	CodeChallengeMethodS512 = CodeChallengeMethod{
		slug: "S512",
		hash: sha512.New(),
	}
)

var ErrCodeChallengeMethodUnknown = errors.New("unknown code challenge method")

//nolint: gochecknoglobals // NOTE(toby3d): maps cannot be constants
var slugsMethods = map[string]CodeChallengeMethod{
	CodeChallengeMethodMD5.slug:   CodeChallengeMethodMD5,
	CodeChallengeMethodPLAIN.slug: CodeChallengeMethodPLAIN,
	CodeChallengeMethodS1.slug:    CodeChallengeMethodS1,
	CodeChallengeMethodS256.slug:  CodeChallengeMethodS256,
	CodeChallengeMethodS512.slug:  CodeChallengeMethodS512,
}

// ParseCodeChallengeMethod parse string identifier of code challenge method
// into struct enum.
func ParseCodeChallengeMethod(slug string) (CodeChallengeMethod, error) {
	if method, ok := slugsMethods[strings.ToUpper(slug)]; ok {
		return method, nil
	}

	return CodeChallengeMethodUndefined, fmt.Errorf("%w: %s", ErrCodeChallengeMethodUnknown, slug)
}

// UnmarshalForm implements custom unmarshler for form values.
func (ccm *CodeChallengeMethod) UnmarshalForm(v []byte) error {
	method, err := ParseCodeChallengeMethod(string(v))
	if err != nil {
		return fmt.Errorf("code_challenge_method: %w", err)
	}

	*ccm = method

	return nil
}

// String returns string representation of code challenge method.
func (ccm CodeChallengeMethod) String() string {
	return ccm.slug
}

func (ccm CodeChallengeMethod) Encoder() hash.Hash {
	return ccm.hash
}
