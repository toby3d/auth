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
		Code       ConfigCode       `yaml:"code"`
		Database   ConfigDatabase   `yaml:"database"`
		IndieAuth  ConfigIndieAuth  `yaml:"indieAuth"`
		JWT        ConfigJWT        `yaml:"jwt"`
		Server     ConfigServer     `yaml:"server"`
		TicketAuth ConfigTicketAuth `yaml:"ticketAuth"`
		Name       string           `yaml:"name"`
		RunMode    string           `yaml:"runMode"`
	}

	ConfigServer struct {
		CertificateFile string `yaml:"certFile"`
		Domain          string `yaml:"domain"`
		Host            string `yaml:"host"`
		KeyFile         string `yaml:"keyFile"`
		Port            string `yaml:"port"`
		Protocol        string `yaml:"protocol"`
		RootURL         string `yaml:"rootUrl"`
		StaticURLPrefix string `yaml:"staticUrlPrefix"`
		EnablePprof     bool   `yaml:"enablePprof"`
	}

	ConfigDatabase struct {
		Path string `yaml:"path"`
		Type string `yaml:"type"` // memory
	}

	// Configuration of a one-time code after giving permission to an
	// application. The client needs to request the server with this code to
	// exchange it for a token or user information.
	ConfigCode struct {
		Expiry time.Duration `yaml:"expiry"` // 10m
		Length uint8         `yaml:"length"` // 32
	}

	ConfigJWT struct {
		Expiry      time.Duration `yaml:"expiry"`    // 1h
		Algorithm   string        `yaml:"algorithm"` // HS256
		Secret      string        `yaml:"secret"`
		NonceLength uint8         `yaml:"nonceLength"` // 22
	}

	ConfigIndieAuth struct {
		Password string `yaml:"password"`
		Username string `yaml:"username"`
		Enabled  bool   `yaml:"enabled"` // true
	}

	ConfigTicketAuth struct {
		Expiry time.Duration `yaml:"expiry"` // 1m
		Length uint8         `yaml:"length"` // 24
	}

	ConfigRelMeAuth struct {
		Providers []ConfigRelMeAuthProvider `yaml:"providers"`
		Enabled   bool                      `yaml:"enabled"` // true
	}

	ConfigRelMeAuthProvider struct {
		ID     string `yaml:"id"`
		Secret string `yaml:"secret"`
		Type   string `yaml:"type"`
	}
)

// TestConfig returns a valid config for tests.
//
//nolint:gomnd // testing domain can contains non-standart values
func TestConfig(tb testing.TB) *Config {
	tb.Helper()

	return &Config{
		Name:    "IndieAuth",
		RunMode: "dev",
		Server: ConfigServer{
			CertificateFile: filepath.Join("https", "cert.pem"),
			Domain:          "localhost",
			EnablePprof:     false,
			Host:            "0.0.0.0",
			KeyFile:         filepath.Join("https", "key.pem"),
			Port:            "3000",
			Protocol:        "http",
			RootURL:         "{{protocol}}://{{domain}}:{{port}}/",
			StaticURLPrefix: "/static",
		},
		Database: ConfigDatabase{
			Type: "memory",
			Path: "",
		},
		Code: ConfigCode{
			Expiry: 10 * time.Minute,
			Length: 32,
		},
		JWT: ConfigJWT{
			Expiry:      time.Hour,
			NonceLength: 22,
			Secret:      "hackme",
			Algorithm:   "HS256",
		},
		IndieAuth: ConfigIndieAuth{
			Enabled:  true,
			Username: "user",
			Password: "password",
		},
		TicketAuth: ConfigTicketAuth{
			Expiry: time.Minute,
			Length: 24,
		},
	}
}

// GetAddress return host:port address.
func (cs ConfigServer) GetAddress() string {
	return net.JoinHostPort(cs.Host, cs.Port)
}

// GetRootURL returns generated root URL from template RootURL.
func (cs ConfigServer) GetRootURL() string {
	return fasttemplate.ExecuteString(cs.RootURL, `{{`, `}}`, map[string]interface{}{
		"domain":          cs.Domain,
		"host":            cs.Host,
		"port":            cs.Port,
		"protocol":        cs.Protocol,
		"staticUrlPrefix": cs.StaticURLPrefix,
	})
}
