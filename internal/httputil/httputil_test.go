package httputil_test

import (
	"testing"

	http "github.com/valyala/fasthttp"

	"source.toby3d.me/toby3d/auth/internal/httputil"
)

const testBody = `<html>
  <body class="h-page">
    <main class="h-card">
      <h1 class="p-name">Sample Name</h1>
    </main>
  </body>
</html>`

func TestExtractProperty(t *testing.T) {
	t.Parallel()

	resp := http.AcquireResponse()
	defer http.ReleaseResponse(resp)
	resp.SetBodyString(testBody)

	results := httputil.ExtractProperty(resp, "h-card", "name")
	if results == nil || results[0] != "Sample Name" {
		t.Errorf(`ExtractProperty(resp, "h-card", "name") = %+s, want %+s`, results, []string{"Sample Name"})
	}
}
