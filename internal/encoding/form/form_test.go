package form_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	http "github.com/valyala/fasthttp"

	"source.toby3d.me/website/oauth/internal/encoding/form"
)

type (
	ResponseType string

	URI struct {
		*http.URI `form:"-"`
	}

	TestResult struct {
		State               []byte       `form:"state"`
		Scope               []string     `form:"scope[]"`
		ClientID            *URI         `form:"client_id"`
		RedirectURI         *URI         `form:"redirect_uri"`
		Me                  *URI         `form:"me"`
		ResponseType        ResponseType `form:"response_type"`
		CodeChallenge       string       `form:"code_challenge"`
		CodeChallengeMethod string       `form:"code_challenge_method"`
	}
)

const testData string = `response_type=code` + // NOTE(toby3d): string type alias
	`&state=1234567890` + // NOTE(toby3d): raw value
	// NOTE(toby3d): custom URL types
	`&client_id=https://app.example.com/` +
	`&redirect_uri=https://app.example.com/redirect` +
	`&me=https://user.example.net/` +
	// NOTE(toby3d): plain strings
	`&code_challenge=OfYAxt8zU2dAPDWQxTAUIteRzMsoj9QBdMIVEDOErUo` +
	`&code_challenge_method=S256` +
	// NOTE(toby3d): multiple values
	`&scope[]=profile` +
	`&scope[]=create` +
	`&scope[]=update` +
	`&scope[]=delete`

func TestUnmarshal(t *testing.T) {
	t.Parallel()

	args := http.AcquireArgs()
	clientId, redirectUri, me := http.AcquireURI(), http.AcquireURI(), http.AcquireURI()

	t.Cleanup(func() {
		http.ReleaseURI(me)
		http.ReleaseURI(redirectUri)
		http.ReleaseURI(clientId)
		http.ReleaseArgs(args)
	})

	require.NoError(t, clientId.Parse(nil, []byte("https://app.example.com/")))
	require.NoError(t, redirectUri.Parse(nil, []byte("https://app.example.com/redirect")))
	require.NoError(t, me.Parse(nil, []byte("https://user.example.net/")))
	args.Parse(testData)

	result := new(TestResult)
	require.NoError(t, form.Unmarshal(args, result))
	assert.Equal(t, &TestResult{
		ClientID:            &URI{URI: clientId},
		Me:                  &URI{URI: me},
		RedirectURI:         &URI{URI: redirectUri},
		State:               []byte("1234567890"),
		Scope:               []string{"profile", "create", "update", "delete"},
		CodeChallengeMethod: "S256",
		CodeChallenge:       "OfYAxt8zU2dAPDWQxTAUIteRzMsoj9QBdMIVEDOErUo",
		ResponseType:        "code",
	}, result)
}

func (src *URI) UnmarshalForm(v []byte) error {
	src.URI = http.AcquireURI()

	return src.Parse(nil, v)
}
