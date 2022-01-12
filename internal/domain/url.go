package domain

import (
	"net/url"
	"strconv"
	"testing"

	http "github.com/valyala/fasthttp"
)

type URL struct {
	*http.URI
}

func NewURL(src string) (*URL, error) {
	u := &URL{
		URI: http.AcquireURI(),
	}
	if err := u.URI.Parse(nil, []byte(src)); err != nil {
		return nil, err
	}

	return u, nil
}

func TestURL(tb testing.TB, src string) *URL {
	tb.Helper()

	u := http.AcquireURI()
	u.Update(src)

	return &URL{
		URI: u,
	}
}

func (u *URL) UnmarshalForm(v []byte) error {
	url, err := NewURL(string(v))
	if err != nil {
		return err
	}

	*u = *url

	return nil
}

func (u *URL) UnmarshalJSON(v []byte) error {
	src, err := strconv.Unquote(string(v))
	if err != nil {
		return err
	}

	url, err := NewURL(src)
	if err != nil {
		return err
	}

	*u = *url

	return nil
}

func (u URL) URL() *url.URL {
	if u.URI == nil {
		return nil
	}

	result, err := url.ParseRequestURI(u.URI.String())
	if err != nil {
		return nil
	}

	return result
}
