package domain_test

import (
	"testing"

	"source.toby3d.me/toby3d/auth/internal/domain"
)

func TestConfigServer_GetAddress(t *testing.T) {
	t.Parallel()

	config := domain.TestConfig(t)
	expResult := config.Server.Host + ":" + config.Server.Port

	if result := config.Server.GetAddress(); result != expResult {
		t.Errorf("GetAddress() = %s, want %s", result, expResult)
	}
}

func TestConfigServer_GetRootURL(t *testing.T) {
	t.Parallel()

	config := domain.TestConfig(t)
	expResult := config.Server.Protocol + "://" + config.Server.Domain + ":" + config.Server.Port + "/"

	if result := config.Server.GetRootURL(); result != expResult {
		t.Errorf("GetRootURL() = %s, want %s", result, expResult)
	}
}
