package util_test

import (
	"testing"

	http "github.com/valyala/fasthttp"

	"source.toby3d.me/website/indieauth/internal/util"
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

	results := util.ExtractProperty(resp, "h-card", "name")
	if results == nil || results[0] != "Sample Name" {
		t.Errorf(`ExtractProperty(resp, "h-card", "name") = %+s, want %+s`, results, []string{"Sample Name"})
	}
}
