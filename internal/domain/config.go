package domain

import (
	"testing"

	"github.com/spf13/viper"
)

// TODO(toby3d): create Config domain

// TestConfig returns a valid *viper.Viper with the generated test data filled in.
func TestConfig(tb testing.TB) *viper.Viper {
	tb.Helper()

	v := viper.New()
	v.Set("indieauth.jwtSecret", "hackme")
	v.Set("indieauth.jwtSigningAlgorithm", "HS256")
	v.Set("server.domain", "example.com")
	v.Set("server.protocol", "https")
	v.Set("server.rootUrl", "{{protocol}}://{{domain}}/")

	return v
}
