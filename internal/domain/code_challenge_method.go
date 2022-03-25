package domain

//nolint: gosec // support old clients
import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"hash"
	"strconv"
	"strings"
)

// CodeChallengeMethod represent a PKCE challenge method for validate verifier.
//
// NOTE(toby3d): Encapsulate enums in structs for extra compile-time safety:
// https://threedots.tech/post/safer-enums-in-go/#struct-based-enums
type CodeChallengeMethod struct {
	hash hash.Hash
	uid  string
}

//nolint: gochecknoglobals // structs cannot be constants
var (
	CodeChallengeMethodUndefined = CodeChallengeMethod{
		uid:  "",
		hash: nil,
	}

	CodeChallengeMethodPLAIN = CodeChallengeMethod{
		uid:  "PLAIN",
		hash: nil,
	}

	CodeChallengeMethodMD5 = CodeChallengeMethod{
		uid:  "MD5",
		hash: md5.New(), //nolint: gosec // support old clients
	}

	CodeChallengeMethodS1 = CodeChallengeMethod{
		uid:  "S1",
		hash: sha1.New(), //nolint: gosec // support old clients
	}

	CodeChallengeMethodS256 = CodeChallengeMethod{
		uid:  "S256",
		hash: sha256.New(),
	}

	CodeChallengeMethodS512 = CodeChallengeMethod{
		uid:  "S512",
		hash: sha512.New(),
	}
)

var ErrCodeChallengeMethodUnknown error = NewError(
	ErrorCodeInvalidRequest,
	"unknown code_challene_method",
	"https://indieauth.net/source/#authorization-request",
)

//nolint: gochecknoglobals // maps cannot be constants
var uidsMethods = map[string]CodeChallengeMethod{
	CodeChallengeMethodMD5.uid:       CodeChallengeMethodMD5,
	CodeChallengeMethodPLAIN.uid:     CodeChallengeMethodPLAIN,
	CodeChallengeMethodS1.uid:        CodeChallengeMethodS1,
	CodeChallengeMethodS256.uid:      CodeChallengeMethodS256,
	CodeChallengeMethodS512.uid:      CodeChallengeMethodS512,
	CodeChallengeMethodUndefined.uid: CodeChallengeMethodUndefined,
}

// ParseCodeChallengeMethod parse string identifier of code challenge method
// into struct enum.
func ParseCodeChallengeMethod(uid string) (CodeChallengeMethod, error) {
	if method, ok := uidsMethods[strings.ToUpper(uid)]; ok {
		return method, nil
	}

	return CodeChallengeMethodUndefined, fmt.Errorf("%w: %s", ErrCodeChallengeMethodUnknown, uid)
}

// UnmarshalForm implements custom unmarshler for form values.
func (ccm *CodeChallengeMethod) UnmarshalForm(v []byte) error {
	method, err := ParseCodeChallengeMethod(string(v))
	if err != nil {
		return fmt.Errorf("CodeChallengeMethod: UnmarshalForm: %w", err)
	}

	*ccm = method

	return nil
}

// UnmarshalJSON implements custom unmarshler for JSON.
func (ccm *CodeChallengeMethod) UnmarshalJSON(v []byte) error {
	src, err := strconv.Unquote(string(v))
	if err != nil {
		return fmt.Errorf("CodeChallengeMethod: UnmarshalJSON: %w", err)
	}

	if *ccm, err = ParseCodeChallengeMethod(src); err != nil {
		return fmt.Errorf("CodeChallengeMethod: UnmarshalJSON: %w", err)
	}

	return nil
}

func (ccm CodeChallengeMethod) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Quote(ccm.uid)), nil
}

// String returns string representation of code challenge method.
func (ccm CodeChallengeMethod) String() string {
	return ccm.uid
}

// Validate checks for a match to the verifier with the hashed version of the
// challenge via the chosen method.
func (ccm CodeChallengeMethod) Validate(codeChallenge, verifier string) bool {
	if ccm.uid == CodeChallengeMethodUndefined.uid {
		return false
	}

	if ccm.uid == CodeChallengeMethodPLAIN.uid {
		return codeChallenge == verifier
	}

	hash := ccm.hash
	hash.Reset() // WARN(toby3d): even hash.New contains something.

	if _, err := hash.Write([]byte(verifier)); err != nil {
		return false
	}

	return codeChallenge == base64.RawURLEncoding.EncodeToString(hash.Sum(nil))
}
