package usecase_test

import (
	"os"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"source.toby3d.me/website/oauth/internal/config"
	repository "source.toby3d.me/website/oauth/internal/config/repository/viper"
	"source.toby3d.me/website/oauth/internal/config/usecase"
)

//nolint: gochecknoglobals
var ucase config.UseCase

func TestMain(m *testing.M) {
	v := viper.New()

	for key, val := range map[string]interface{}{
		"database.client":              "bolt",
		"database.connection.filename": "./data/development.db",
		"server.host":                  "127.0.0.1",
		"server.port":                  3000,
		"url":                          "http://127.0.0.1:3000/",
	} {
		v.Set(key, val)
	}

	ucase = usecase.NewConfigUseCase(repository.NewViperConfigRepository(v))

	os.Exit(m.Run())
}

func TestAddr(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "127.0.0.1:3000", ucase.Addr())
}

func TestDBFileName(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "./data/development.db", ucase.DBFileName())
}

func TestHost(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "127.0.0.1", ucase.Host())
}

func TestPort(t *testing.T) {
	t.Parallel()

	assert.Equal(t, 3000, ucase.Port())
}

func TestURL(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "http://127.0.0.1:3000/", ucase.URL())
}
