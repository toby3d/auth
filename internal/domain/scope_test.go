package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"source.toby3d.me/website/indieauth/internal/domain"
)

func TestScopesUnmarshalJSON(t *testing.T) {
	t.Parallel()

	result := &struct {
		Scope domain.Scopes `json:"scope"`
	}{}
	require.NoError(t, json.Unmarshal([]byte(`{"scope": "read update delete"}`), result))

	for _, scope := range []domain.Scope{domain.ScopeRead, domain.ScopeUpdate, domain.ScopeDelete} {
		assert.Contains(t, result.Scope, scope)
	}
}
