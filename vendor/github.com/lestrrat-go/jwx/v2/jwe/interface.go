package jwe

import (
	"github.com/lestrrat-go/iter/mapiter"
	"github.com/lestrrat-go/jwx/v2/internal/iter"
	"github.com/lestrrat-go/jwx/v2/jwe/internal/keygen"
)

// Recipient holds the encrypted key and hints to decrypt the key
type Recipient interface {
	Headers() Headers
	EncryptedKey() []byte
	SetHeaders(Headers) error
	SetEncryptedKey([]byte) error
}

type stdRecipient struct {
	// Comments on each field are taken from https://datatracker.ietf.org/doc/html/rfc7516
	//
	// header
	//    The "header" member MUST be present and contain the value JWE Per-
	//    Recipient Unprotected Header when the JWE Per-Recipient
	//    Unprotected Header value is non-empty; otherwise, it MUST be
	//    absent.  This value is represented as an unencoded JSON object,
	//    rather than as a string.  These Header Parameter values are not
	//    integrity protected.
	//
	// At least one of the "header", "protected", and "unprotected" members
	// MUST be present so that "alg" and "enc" Header Parameter values are
	// conveyed for each recipient computation.
	//
	// JWX note: see Message.unprotectedHeaders
	headers Headers

	// encrypted_key
	//    The "encrypted_key" member MUST be present and contain the value
	//    BASE64URL(JWE Encrypted Key) when the JWE Encrypted Key value is
	//    non-empty; otherwise, it MUST be absent.
	encryptedKey []byte
}

// Message contains the entire encrypted JWE message. You should not
// expect to use Message for anything other than inspecting the
// state of an encrypted message. This is because encryption is
// highly context sensitive, and once we parse the original payload
// into an object, we may not always be able to recreate the exact
// context in which the encryption happened.
//
// For example, it is totally valid for if the protected header's
// integrity was calculated using a non-standard line breaks:
//
//    {"a dummy":
//      "protected header"}
//
// Once parsed, though, we can only serialize the protected header as:
//
//    {"a dummy":"protected header"}
//
// which would obviously result in a contradicting integrity value
// if we tried to re-calculate it from a parsed message.
//nolint:govet
type Message struct {
	// Comments on each field are taken from https://datatracker.ietf.org/doc/html/rfc7516
	//
	// protected
	//    The "protected" member MUST be present and contain the value
	//    BASE64URL(UTF8(JWE Protected Header)) when the JWE Protected
	//    Header value is non-empty; otherwise, it MUST be absent.  These
	//    Header Parameter values are integrity protected.
	protectedHeaders Headers

	// unprotected
	//    The "unprotected" member MUST be present and contain the value JWE
	//    Shared Unprotected Header when the JWE Shared Unprotected Header
	//    value is non-empty; otherwise, it MUST be absent.  This value is
	//    represented as an unencoded JSON object, rather than as a string.
	//    These Header Parameter values are not integrity protected.
	//
	// JWX note: This field is NOT mutually exclusive with per-recipient
	// headers within the implmentation because... it's too much work.
	// It is _never_ populated (we don't provide a way to do this) upon encryption.
	// When decrypting, if present its values are always merged with
	// per-recipient header.
	unprotectedHeaders Headers

	// iv
	//    The "iv" member MUST be present and contain the value
	//    BASE64URL(JWE Initialization Vector) when the JWE Initialization
	//    Vector value is non-empty; otherwise, it MUST be absent.
	initializationVector []byte

	// aad
	//    The "aad" member MUST be present and contain the value
	//    BASE64URL(JWE AAD)) when the JWE AAD value is non-empty;
	//    otherwise, it MUST be absent.  A JWE AAD value can be included to
	//    supply a base64url-encoded value to be integrity protected but not
	//    encrypted.
	authenticatedData []byte

	// ciphertext
	//    The "ciphertext" member MUST be present and contain the value
	//    BASE64URL(JWE Ciphertext).
	cipherText []byte

	// tag
	//    The "tag" member MUST be present and contain the value
	//    BASE64URL(JWE Authentication Tag) when the JWE Authentication Tag
	//    value is non-empty; otherwise, it MUST be absent.
	tag []byte

	// recipients
	//    The "recipients" member value MUST be an array of JSON objects.
	//    Each object contains information specific to a single recipient.
	//    This member MUST be present with exactly one array element per
	//    recipient, even if some or all of the array element values are the
	//    empty JSON object "{}" (which can happen when all Header Parameter
	//    values are shared between all recipients and when no encrypted key
	//    is used, such as when doing Direct Encryption).
	//
	// Some Header Parameters, including the "alg" parameter, can be shared
	// among all recipient computations.  Header Parameters in the JWE
	// Protected Header and JWE Shared Unprotected Header values are shared
	// among all recipients.
	//
	// The Header Parameter values used when creating or validating per-
	// recipient ciphertext and Authentication Tag values are the union of
	// the three sets of Header Parameter values that may be present: (1)
	// the JWE Protected Header represented in the "protected" member, (2)
	// the JWE Shared Unprotected Header represented in the "unprotected"
	// member, and (3) the JWE Per-Recipient Unprotected Header represented
	// in the "header" member of the recipient's array element.  The union
	// of these sets of Header Parameters comprises the JOSE Header.  The
	// Header Parameter names in the three locations MUST be disjoint.
	recipients []Recipient

	// TODO: Additional members can be present in both the JSON objects defined
	// above; if not understood by implementations encountering them, they
	// MUST be ignored.
	// privateParams map[string]interface{}

	// These two fields below are not available for the public consumers of this object.
	// rawProtectedHeaders stores the original protected header buffer
	rawProtectedHeaders []byte
	// storeProtectedHeaders is a hint to be used in UnmarshalJSON().
	// When this flag is true, UnmarshalJSON() will populate the
	// rawProtectedHeaders field
	storeProtectedHeaders bool
}

// populater is an interface for things that may modify the
// JWE header. e.g. ByteWithECPrivateKey
type populater interface {
	Populate(keygen.Setter) error
}

type Visitor = iter.MapVisitor
type VisitorFunc = iter.MapVisitorFunc
type HeaderPair = mapiter.Pair
type Iterator = mapiter.Iterator
