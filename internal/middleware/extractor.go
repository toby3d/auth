package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"net/textproto"
	"strings"

	"source.toby3d.me/toby3d/auth/internal/common"
)

// ValuesExtractor defines a function for extracting values (keys/tokens) from the given context.
type ValuesExtractor func(w http.ResponseWriter, r *http.Request) ([]string, error)

// extractorLimit is arbitrary number to limit values extractor can return. this limits possible resource
// exhaustion attack vector.
const extractorLimit = 20

var (
	errHeaderExtractorValueMissing = errors.New("missing value in request header")
	errHeaderExtractorValueInvalid = errors.New("invalid value in request header")
	errQueryExtractorValueMissing  = errors.New("missing value in the query string")
	errCookieExtractorValueMissing = errors.New("missing value in cookies")
	errFormExtractorValueMissing   = errors.New("missing value in the form")
)

// CreateExtractors creates ValuesExtractors from given lookups.
// Lookups is a string in the form of "<source>:<name>" or "<source>:<name>,<source>:<name>" that is used
// to extract key from the request.
// Possible values:
//   - "header:<name>" or "header:<name>:<cut-prefix>"
//     `<cut-prefix>` is argument value to cut/trim prefix of the extracted value. This is useful if header
//     value has static prefix like `Authorization: <auth-scheme> <authorisation-parameters>` where part that we
//     want to cut is `<auth-scheme> ` note the space at the end.
//     In case of basic authentication `Authorization: Basic <credentials>` prefix we want to remove is `Basic `.
//   - "query:<name>"
//   - "param:<name>"
//   - "form:<name>"
//   - "cookie:<name>"
//
// Multiple sources example:
// - "header:Authorization,header:X-Api-Key".
func CreateExtractors(lookups string) ([]ValuesExtractor, error) {
	return createExtractors(lookups, "")
}

func createExtractors(lookups, authScheme string) ([]ValuesExtractor, error) {
	if lookups == "" {
		return nil, nil
	}

	sources := strings.Split(lookups, ",")
	extractors := make([]ValuesExtractor, 0)

	for _, source := range sources {
		parts := strings.Split(source, ":")

		if len(parts) <= 1 {
			return nil, fmt.Errorf("extractor source for lookup could not be split into needed parts: %v",
				source)
		}

		switch parts[0] {
		case "query":
			extractors = append(extractors, valuesFromQuery(parts[1]))
		case "cookie":
			extractors = append(extractors, valuesFromCookie(parts[1]))
		case "form":
			extractors = append(extractors, valuesFromForm(parts[1]))
		case "header":
			prefix := ""
			if len(parts) > 2 {
				prefix = parts[2]
			} else if authScheme != "" && parts[1] == common.HeaderAuthorization {
				// backwards compatibility for JWT and KeyAuth:
				// * we only apply this fix to Authorization as header we use and uses prefixes like
				//   "Bearer <token-value>" etc
				// * previously header extractor assumed that auth-scheme/prefix had a space as suffix
				//   we need to retain that behaviour for default values and Authorization header.
				prefix = authScheme
				if !strings.HasSuffix(prefix, " ") {
					prefix += " "
				}
			}

			extractors = append(extractors, valuesFromHeader(parts[1], prefix))
		}
	}

	return extractors, nil
}

// valuesFromHeader returns a functions that extracts values from the request header.
// valuePrefix is parameter to remove first part (prefix) of the extracted value. This is useful if header value has
// static prefix like `Authorization: <auth-scheme> <authorisation-parameters>` where part that we want to remove is
// `<auth-scheme> ` note the space at the end. In case of basic authentication `Authorization: Basic <credentials>`
// prefix we want to remove is `Basic `. In case of JWT tokens `Authorization: Bearer <token>` prefix is `Bearer `.
// If prefix is left empty the whole value is returned.
func valuesFromHeader(header, valuePrefix string) ValuesExtractor {
	prefixLen := len(valuePrefix)
	// standard library parses http.Request header keys in canonical form but we may provide something else so fix
	// this
	header = textproto.CanonicalMIMEHeaderKey(header)

	return func(_ http.ResponseWriter, r *http.Request) ([]string, error) {
		values := r.Header.Values(header)
		if len(values) == 0 {
			return nil, errHeaderExtractorValueMissing
		}

		result := make([]string, 0)

		for i, value := range values {
			if prefixLen == 0 {
				result = append(result, value)

				if i >= extractorLimit-1 {
					break
				}

				continue
			}

			if len(value) > prefixLen && strings.EqualFold(value[:prefixLen], valuePrefix) {
				result = append(result, value[prefixLen:])

				if i >= extractorLimit-1 {
					break
				}
			}
		}

		if len(result) == 0 {
			if prefixLen > 0 {
				return nil, errHeaderExtractorValueInvalid
			}

			return nil, errHeaderExtractorValueMissing
		}

		return result, nil
	}
}

// valuesFromQuery returns a function that extracts values from the query string.
func valuesFromQuery(param string) ValuesExtractor {
	return func(_ http.ResponseWriter, r *http.Request) ([]string, error) {
		result := r.URL.Query()[param]

		if len(result) == 0 {
			return nil, errQueryExtractorValueMissing
		} else if len(result) > extractorLimit-1 {
			result = result[:extractorLimit]
		}

		return result, nil
	}
}

// valuesFromCookie returns a function that extracts values from the named cookie.
func valuesFromCookie(name string) ValuesExtractor {
	return func(_ http.ResponseWriter, r *http.Request) ([]string, error) {
		cookies := r.Cookies()
		if len(cookies) == 0 {
			return nil, errCookieExtractorValueMissing
		}

		result := make([]string, 0)

		for i, cookie := range cookies {
			if name == cookie.Name {
				result = append(result, cookie.Value)

				if i >= extractorLimit-1 {
					break
				}
			}
		}

		if len(result) == 0 {
			return nil, errCookieExtractorValueMissing
		}

		return result, nil
	}
}

// valuesFromForm returns a function that extracts values from the form field.
func valuesFromForm(name string) ValuesExtractor {
	return func(_ http.ResponseWriter, r *http.Request) ([]string, error) {
		if r.Form == nil {
			_ = r.ParseMultipartForm(32 << 20) // same what `r.FormValue(name)` does
		}

		values := r.Form[name]

		if len(values) == 0 {
			return nil, errFormExtractorValueMissing
		}

		if len(values) > extractorLimit-1 {
			values = values[:extractorLimit]
		}

		result := append([]string{}, values...)

		return result, nil
	}
}
