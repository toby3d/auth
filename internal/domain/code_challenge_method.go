package domain

//nolint:gosec // support old clients
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

	"source.toby3d.me/toby3d/auth/internal/common"
)

// CodeChallengeMethod represent a PKCE challenge method for validate verifier.
//
// NOTE(toby3d): Encapsulate enums in structs for extra compile-time safety:
// https://threedots.tech/post/safer-enums-in-go/#struct-based-enums
type CodeChallengeMethod struct {
	uid string
}

//nolint:gochecknoglobals // structs cannot be constants
var (
	CodeChallengeMethodUnd   = CodeChallengeMethod{uid: ""}      // "und"
	CodeChallengeMethodPLAIN = CodeChallengeMethod{uid: "plain"} // "PLAIN"
	CodeChallengeMethodMD5   = CodeChallengeMethod{uid: "md5"}   // "MD5"
	CodeChallengeMethodS1    = CodeChallengeMethod{uid: "s1"}    // "S1"
	CodeChallengeMethodS256  = CodeChallengeMethod{uid: "s256"}  // "S256"
	CodeChallengeMethodS512  = CodeChallengeMethod{uid: "s512"}  // "S512"
)

var ErrCodeChallengeMethodUnknown error = NewError(
	ErrorCodeInvalidRequest,
	"unknown code_challene_method",
	"https://indieauth.net/source/#authorization-request",
)

//nolint:gochecknoglobals // maps cannot be constants
var uidsMethods = map[string]CodeChallengeMethod{
	CodeChallengeMethodMD5.uid:   CodeChallengeMethodMD5,
	CodeChallengeMethodPLAIN.uid: CodeChallengeMethodPLAIN,
	CodeChallengeMethodS1.uid:    CodeChallengeMethodS1,
	CodeChallengeMethodS256.uid:  CodeChallengeMethodS256,
	CodeChallengeMethodS512.uid:  CodeChallengeMethodS512,
}

// ParseCodeChallengeMethod parse string identifier of code challenge method
// into struct enum.
func ParseCodeChallengeMethod(uid string) (CodeChallengeMethod, error) {
	if method, ok := uidsMethods[strings.ToLower(uid)]; ok {
		return method, nil
	}

	return CodeChallengeMethodUnd, fmt.Errorf("%w: %s", ErrCodeChallengeMethodUnknown, uid)
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
	if ccm.uid != "" {
		return strings.ToUpper(ccm.uid)
	}

	return common.Und
}

func (ccm CodeChallengeMethod) GoString() string {
	return "domain.CodeChallengeMethod(" + ccm.String() + ")"
}

// Validate checks for a match to the verifier with the hashed version of the
// challenge via the chosen method.
func (ccm CodeChallengeMethod) Validate(codeChallenge, verifier string) bool {
	var h hash.Hash

	switch ccm {
	default:
		return false
	case CodeChallengeMethodPLAIN:
		return codeChallenge == verifier
	case CodeChallengeMethodMD5:
		h = md5.New()
	case CodeChallengeMethodS1:
		h = sha1.New()
	case CodeChallengeMethodS256:
		h = sha256.New()
	case CodeChallengeMethodS512:
		h = sha512.New()
	}

	if _, err := h.Write([]byte(verifier)); err != nil {
		return false
	}

	return codeChallenge == base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}
