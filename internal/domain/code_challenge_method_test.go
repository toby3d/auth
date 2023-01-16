package domain_test

//nolint:gosec // support old clients
import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"hash"
	"testing"

	"github.com/brianvoe/gofakeit/v6"

	"source.toby3d.me/toby3d/auth/internal/domain"
	"source.toby3d.me/toby3d/auth/internal/random"
)

func TestParseCodeChallengeMethod(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name     string
		in       string
		out      domain.CodeChallengeMethod
		expError bool
	}{
		{name: "invalid", in: "und", out: domain.CodeChallengeMethodUnd, expError: true},
		{name: "PLAIN", in: "plain", out: domain.CodeChallengeMethodPLAIN, expError: false},
		{name: "MD5", in: "Md5", out: domain.CodeChallengeMethodMD5, expError: false},
		{name: "S1", in: "S1", out: domain.CodeChallengeMethodS1, expError: false},
		{name: "S256", in: "s256", out: domain.CodeChallengeMethodS256, expError: false},
		{name: "S512", in: "S512", out: domain.CodeChallengeMethodS512, expError: false},
	} {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result, err := domain.ParseCodeChallengeMethod(tc.in)

			switch {
			case err != nil && !tc.expError:
				t.Errorf("ParseCodeChallengeMethod(%s) = %+v, want nil", tc.in, err)
			case err == nil && tc.expError:
				t.Errorf("ParseCodeChallengeMethod(%s) = %+v, want error", tc.in, err)
			}

			if result != tc.out {
				t.Errorf("ParseCodeChallengeMethod(%s) = %v, want %v", tc.in, result, tc.out)
			}
		})
	}
}

func TestCodeChallengeMethod_UnmarshalForm(t *testing.T) {
	t.Parallel()

	input := []byte("S256")
	result := domain.CodeChallengeMethodUnd

	if err := result.UnmarshalForm(input); err != nil {
		t.Fatalf("%+v", err)
	}

	if result != domain.CodeChallengeMethodS256 {
		t.Errorf("UnmarshalForm(%s) = %v, want %v", input, result, domain.CodeChallengeMethodS256)
	}
}

func TestCodeChallengeMethod_UnmarshalJSON(t *testing.T) {
	t.Parallel()

	input := []byte(`"S256"`)
	result := domain.CodeChallengeMethodUnd

	if err := result.UnmarshalJSON(input); err != nil {
		t.Fatalf("%+v", err)
	}

	if result != domain.CodeChallengeMethodS256 {
		t.Errorf("UnmarshalJSON(%s) = %v, want %v", input, result, domain.CodeChallengeMethodS256)
	}
}

func TestCodeChallengeMethod_String(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		in   domain.CodeChallengeMethod
		out  string
	}{
		{name: "plain", in: domain.CodeChallengeMethodPLAIN, out: "PLAIN"},
		{name: "md5", in: domain.CodeChallengeMethodMD5, out: "MD5"},
		{name: "s1", in: domain.CodeChallengeMethodS1, out: "S1"},
		{name: "s256", in: domain.CodeChallengeMethodS256, out: "S256"},
		{name: "s512", in: domain.CodeChallengeMethodS512, out: "S512"},
	} {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := tc.in.String()
			if result != tc.out {
				t.Errorf("String() = %v, want %v", result, tc.out)
			}
		})
	}
}

//nolint:gosec // support old clients
func TestCodeChallengeMethod_Validate(t *testing.T) {
	t.Parallel()

	verifier, err := random.String(uint8(gofakeit.Number(43, 128)))
	if err != nil {
		t.Fatalf("%+v", err)
	}

	for _, tc := range []struct {
		hash     hash.Hash
		in       domain.CodeChallengeMethod
		name     string
		expError bool
	}{
		{name: "invalid", in: domain.CodeChallengeMethodS256, hash: md5.New(), expError: true},
		{name: "MD5", in: domain.CodeChallengeMethodMD5, hash: md5.New(), expError: false},
		{name: "plain", in: domain.CodeChallengeMethodPLAIN, hash: nil, expError: false},
		{name: "S1", in: domain.CodeChallengeMethodS1, hash: sha1.New(), expError: false},
		{name: "S256", in: domain.CodeChallengeMethodS256, hash: sha256.New(), expError: false},
		{name: "S512", in: domain.CodeChallengeMethodS512, hash: sha512.New(), expError: false},
		{name: "Und", in: domain.CodeChallengeMethodUnd, hash: nil, expError: true},
	} {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var codeChallenge string

			switch tc.in {
			case domain.CodeChallengeMethodUnd, domain.CodeChallengeMethodPLAIN:
				codeChallenge = verifier
			default:
				hash := tc.hash
				hash.Reset()

				if _, err := hash.Write([]byte(verifier)); err != nil {
					t.Error(err)
				}

				codeChallenge = base64.RawURLEncoding.EncodeToString(hash.Sum(nil))
			}

			if result := tc.in.Validate(codeChallenge, verifier); result != !tc.expError {
				t.Errorf("Validate(%s, %s) = %t, want %t", codeChallenge, verifier, result, tc.expError)
			}
		})
	}
}

func TestCodeChallengeMethod_Validate_IndieAuth(t *testing.T) {
	t.Parallel()

	if ok := domain.CodeChallengeMethodS256.Validate(
		"ALiMNf5FvF_LIWLhSkd9tjPKh3PEmai2OrdDBzrVZ3M",
		"6f535c952339f0670311b4bbec5c41c00805e83291fc7eb15ca4963f82a4d57595787dcc6ee90571fb7789cbd521fe0178ed",
	); !ok {
		t.Errorf("Validate(%s, %s) = %t, want %t", "ALiMNf5FvF_LIWLhSkd9tjPKh3PEmai2OrdDBzrVZ3M",
			"6f535c952339f0670311b4bbec5c41c00805e83291fc7eb15ca4963f82a4d57595787dcc6ee90571fb7789cbd521"+
				"fe0178ed", ok, true)
	}
}
