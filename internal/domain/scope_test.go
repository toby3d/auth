package domain_test

import (
	"testing"

	"github.com/goccy/go-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"source.toby3d.me/website/indieauth/internal/domain"
)

/* TODO(toby3d): enable this after form package patch
func TestScopes_UnmarshalForm(t *testing.T) {
	t.Parallel()

	args := http.AcquireArgs()
	defer http.ReleaseArgs(args)
	args.Set("scope", "read update delete")

	result := struct {
		Scope domain.Scopes
	}{
		Scope: make(domain.Scopes, 0),
	}

	require.NoError(t, form.Unmarshal(args, &result))
	assert.Equal(t, "read update delete", result.Scope.String())
}
*/

func TestScopes_UnmarshalJSON(t *testing.T) {
	t.Parallel()

	result := struct {
		Scope domain.Scopes `json:"scope"`
	}{}
	require.NoError(t, json.Unmarshal([]byte(`{"scope":"read update delete"}`), &result))
	assert.Equal(t, domain.Scopes{domain.ScopeRead, domain.ScopeUpdate, domain.ScopeDelete}, result.Scope)
}

func TestScopes_MarshalJSON(t *testing.T) {
	t.Parallel()

	result, err := json.Marshal(map[string]domain.Scopes{
		"scope": {
			domain.ScopeEmail,
			domain.ScopeProfile,
			domain.ScopeRead,
		},
	})
	require.NoError(t, err)
	assert.Equal(t, `{"scope":"email profile read"}`, string(result))
}
