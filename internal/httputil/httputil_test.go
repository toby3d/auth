package httputil_test

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"source.toby3d.me/toby3d/auth/internal/httputil"
)

const testBody = `<html>
  <head>
    <link rel="lipsum" href="https://example.com/">
    <link rel="lipsum" href="https://example.net/">
  </head>
  <body class="h-page">
    <main class="h-app">
      <h1 class="p-name">Sample Name</h1>
    </main>
  </body>
</html>`

func TestExtractEndpointsFromBody(t *testing.T) {
	t.Parallel()

	req, err := http.NewRequest(http.MethodGet, "https://example.com/", nil)
	if err != nil {
		t.Fatal(err)
	}

	in := &http.Response{
		Body:    ioutil.NopCloser(strings.NewReader(testBody)),
		Request: req,
	}

	out, err := httputil.ExtractEndpointsFromBody(in.Body, req.URL, "lipsum")
	if err != nil {
		t.Fatal(err)
	}

	exp := []*url.URL{
		{Scheme: "https", Host: "example.com", Path: "/"},
		{Scheme: "https", Host: "example.net", Path: "/"},
	}

	if !cmp.Equal(out, exp) {
		t.Errorf(`ExtractProperty(resp, "h-card", "name") = %+s, want %+s`, out, exp)
	}
}

func TestExtractProperty(t *testing.T) {
	t.Parallel()

	req, err := http.NewRequest(http.MethodGet, "https://example.com/", nil)
	if err != nil {
		t.Fatal(err)
	}

	in := &http.Response{
		Body:    ioutil.NopCloser(strings.NewReader(testBody)),
		Request: req,
	}

	if out := httputil.ExtractProperty(in.Body, req.URL, "h-app", "name"); out == nil || out[0] != "Sample Name" {
		t.Errorf(`ExtractProperty(%s, %s, %s) = %+s, want %+s`, req.URL, "h-app", "name", out,
			[]string{"Sample Name"})
	}
}
