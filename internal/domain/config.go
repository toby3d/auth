package domain

import (
	"net"
	"path/filepath"
	"testing"
	"time"

	"github.com/valyala/fasttemplate"
)

type (
	Config struct {
		Database  ConfigDatabase
		IndieAuth ConfigIndieAuth
		Server    ConfigServer
		Name      string
		RunMode   string
	}

	ConfigIndieAuth struct {
		JWTSecret                 interface{}
		AccessTokenExpirationTime time.Duration
		JWTSigningAlgorithm       string
		JWTSigningPrivateKeyFile  string
		CodeLength                int
		JWTNonceLength            int
		Enabled                   bool
	}

	ConfigServer struct {
		CertificateFile string
		Domain          string
		Host            string
		KeyFile         string
		Protocol        string
		RootURL         string
		StaticRootPath  string
		StaticURLPrefix string
		Port            string
		EnablePprof     bool
	}

	ConfigDatabase struct {
		Path string
		Type string
	}
)

// GetAddress return host:port address.
func (cs *ConfigServer) GetAddress() string {
	return net.JoinHostPort(cs.Host, cs.Port)
}

// GetRootURL returns generated from template RootURL.
func (cs *ConfigServer) GetRootURL() string {
	return fasttemplate.ExecuteString(cs.RootURL, `{{`, `}}`, map[string]interface{}{
		"domain":          cs.Domain,
		"host":            cs.Host,
		"port":            cs.Port,
		"protocol":        cs.Protocol,
		"staticRootPath":  cs.StaticRootPath,
		"staticUrlPrefix": cs.StaticURLPrefix,
	})
}

// TestConfig returns a valid *viper.Viper with the generated test data filled in.
func TestConfig(tb testing.TB) *Config {
	tb.Helper()

	return &Config{
		Name:    "IndieAuth",
		RunMode: "dev",
		Database: ConfigDatabase{
			Path: filepath.Join("test", "development.db"),
			Type: "bolt",
		},
		IndieAuth: ConfigIndieAuth{
			AccessTokenExpirationTime: time.Hour,
			CodeLength:                32, //nolint: gomnd
			Enabled:                   true,
			JWTNonceLength:            22, //nolint: gomnd
			JWTSecret:                 []byte("hackme"),
			JWTSigningAlgorithm:       "HS256",
			JWTSigningPrivateKeyFile:  filepath.Join("jwt", "private.pem"),
		},
		Server: ConfigServer{
			CertificateFile: filepath.Join("https", "cert.pem"),
			Domain:          "localhost",
			EnablePprof:     false,
			Host:            "0.0.0.0",
			KeyFile:         filepath.Join("https", "key.pem"),
			Port:            "3000",
			Protocol:        "http",
			RootURL:         "{{protocol}}://{{domain}}:{{port}}/",
			StaticRootPath:  "/",
			StaticURLPrefix: "/static",
		},
	}
}
