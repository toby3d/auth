package middleware

import (
	"bytes"
	"errors"
	"fmt"
	"net/textproto"
	"strings"

	http "github.com/valyala/fasthttp"
)

// ValuesExtractor defines a function for extracting values (keys/tokens) from
// the given context.
type ValuesExtractor func(ctx *http.RequestCtx) ([][]byte, error)

const (
	// extractorLimit is arbitrary number to limit values extractor can
	// return. this limits possible resource exhaustion attack vector.
	extractorLimit = 20
)

var (
	errHeaderExtractorValueMissing = errors.New("missing value in request header")
	errHeaderExtractorValueInvalid = errors.New("invalid value in request header")
	errQueryExtractorValueMissing  = errors.New("missing value in the query string")
	errParamExtractorValueMissing  = errors.New("missing value in path params")
	errCookieExtractorValueMissing = errors.New("missing value in cookies")
	errFormExtractorValueMissing   = errors.New("missing value in the form")
)

func createExtractors(lookups, authScheme string) ([]ValuesExtractor, error) {
	if lookups == "" {
		return nil, nil
	}

	sources := strings.Split(lookups, ",")
	extractors := make([]ValuesExtractor, 0)

	for _, source := range sources {
		parts := strings.Split(source, ":")

		if len(parts) < 2 {
			return nil, fmt.Errorf("extractor source for lookup could not be split into needed parts: %s",
				source)
		}

		switch parts[0] {
		case "query":
			extractors = append(extractors, valuesFromQuery(parts[1]))
		case "param":
			extractors = append(extractors, valuesFromParam(parts[1]))
		case "cookie":
			extractors = append(extractors, valuesFromCookie(parts[1]))
		case "form":
			extractors = append(extractors, valuesFromForm(parts[1]))
		case "header":
			prefix := ""

			if len(parts) > 2 {
				prefix = parts[2]
			} else if authScheme != "" && parts[1] == http.HeaderAuthorization {
				// backwards compatibility for JWT and KeyAuth:
				// * we only apply this fix to Authorization as header we use and uses prefixes like "Bearer <token-value>" etc
				// * previously header extractor assumed that auth-scheme/prefix had a space as suffix we need to retain that
				//   behaviour for default values and Authorization header.
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
// valuePrefix is parameter to remove first part (prefix) of the extracted value. This is useful if header value has static
// prefix like `Authorization: <auth-scheme> <authorisation-parameters>` where part that we want to remove is `<auth-scheme> `
// note the space at the end. In case of basic authentication `Authorization: Basic <credentials>` prefix we want to remove
// is `Basic `. In case of JWT tokens `Authorization: Bearer <token>` prefix is `Bearer `.
// If prefix is left empty the whole value is returned.
func valuesFromHeader(header, valuePrefix string) ValuesExtractor {
	prefixLen := len(valuePrefix)
	// standard library parses http.Request header keys in canonical form but we may provide something else so fix this
	header = textproto.CanonicalMIMEHeaderKey(header)

	return func(ctx *http.RequestCtx) ([][]byte, error) {
		value := ctx.Request.Header.Peek(header)
		if len(value) == 0 {
			return nil, errHeaderExtractorValueMissing
		}

		if prefixLen == 0 {
			return append([][]byte{}, value), nil
		}

		if len(value) > prefixLen && bytes.EqualFold(value[:prefixLen], []byte(valuePrefix)) {
			return append([][]byte{}, value[prefixLen:]), nil
		}

		return nil, errHeaderExtractorValueInvalid
	}
}

// valuesFromQuery returns a function that extracts values from the query string.
func valuesFromQuery(param string) ValuesExtractor {
	return func(ctx *http.RequestCtx) ([][]byte, error) {
		if !ctx.QueryArgs().Has(param) {
			return nil, errQueryExtractorValueMissing
		}

		result := ctx.QueryArgs().PeekMulti(param)
		if len(result) > extractorLimit-1 {
			result = result[:extractorLimit]
		}

		return result, nil
	}
}

// valuesFromParam returns a function that extracts values from the url param string.
func valuesFromParam(param string) ValuesExtractor {
	return func(ctx *http.RequestCtx) ([][]byte, error) {
		if !ctx.PostArgs().Has(param) {
			return nil, errParamExtractorValueMissing
		}

		result := ctx.PostArgs().PeekMulti(param)
		if len(result) > extractorLimit-1 {
			result = result[:extractorLimit]
		}

		return result, nil
	}
}

// valuesFromCookie returns a function that extracts values from the named cookie.
func valuesFromCookie(name string) ValuesExtractor {
	return func(ctx *http.RequestCtx) ([][]byte, error) {
		if value := ctx.Request.Header.Cookie(name); len(value) != 0 {
			return append([][]byte{}, value), nil
		}

		return nil, errCookieExtractorValueMissing
	}
}

// valuesFromForm returns a function that extracts values from the form field.
func valuesFromForm(name string) ValuesExtractor {
	return func(ctx *http.RequestCtx) ([][]byte, error) {
		form, err := ctx.MultipartForm()
		if err != nil {
			return nil, fmt.Errorf("valuesFromForm parse form failed: %w", err)
		}

		values := form.Value[name]

		switch {
		case len(values) == 0:
			return nil, errFormExtractorValueMissing
		case len(values) > extractorLimit-1:
			values = values[:extractorLimit]
		}

		result := make([][]byte, 0)
		for i := range values {
			result = append(result, []byte(values[i]))
		}

		return result, nil
	}
}
