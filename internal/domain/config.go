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
		Code       ConfigCode       `envPrefix:"CODE_"`
		Database   ConfigDatabase   `envPrefix:"DATABASE_"`
		IndieAuth  ConfigIndieAuth  `envPrefix:"INDIEAUTH_"`
		JWT        ConfigJWT        `envPrefix:"JWT_"`
		Server     ConfigServer     `envPrefix:"SERVER_"`
		TicketAuth ConfigTicketAuth `envPrefix:"TICKETAUTH_"`
		Name       string           `env:"NAME" envDefault:"IndieAuth"`
		RunMode    string           `env:"RUN_MODE" envDefault:"dev"`
	}

	ConfigServer struct {
		CertificateFile string `env:"CERT_FILE"`
		Domain          string `env:"DOMAIN" envDefault:"localhost"`
		Host            string `env:"HOST" envDefault:"0.0.0.0"`
		KeyFile         string `env:"KEY_FILE"`
		Port            string `env:"PORT" envDefault:"3000"`
		Protocol        string `env:"PROTOCOL" envDefault:"http"`
		RootURL         string `env:"ROOT_URL" envDefault:"{{protocol}}://{{domain}}:{{port}}/"`
		StaticURLPrefix string `env:"STATIC_URL_PREFIX"`
		EnablePprof     bool   `env:"ENABLE_PPROF"`
	}

	ConfigDatabase struct {
		Path string `env:"PATH"`
		Type string `env:"TYPE" envDefault:"memory"` // memory
	}

	// Configuration of a one-time code after giving permission to an
	// application. The client needs to request the server with this code to
	// exchange it for a token or user information.
	ConfigCode struct {
		Expiry time.Duration `env:"EXPIRY" envDefault:"10m"` // 10m
		Length uint8         `env:"LENGTH" envDefault:"32"`  // 32
	}

	ConfigJWT struct {
		Expiry      time.Duration `env:"EXPIRY" envDefault:"1h"`       // 1h
		Algorithm   string        `env:"ALGORITHM" envDefault:"HS256"` // HS256
		Secret      string        `env:"SECRET"`
		NonceLength uint8         `env:"NONCE_LENGTH" envDefault:"22"` // 22
	}

	ConfigIndieAuth struct {
		Password string `env:"PASSWORD"`
		Username string `env:"USERNAME"`
		Enabled  bool   `env:"ENABLED" envDefault:"true"` // true
	}

	ConfigTicketAuth struct {
		Expiry time.Duration `env:"EXPIRY" envDefault:"1m"` // 1m
		Length uint8         `env:"LENGTH" envDefault:"24"` // 24
	}

	ConfigRelMeAuth struct {
		Providers []ConfigRelMeAuthProvider `envPrefix:"PROVIDERS_"`
		Enabled   bool                      `env:"ENABLED" envDefault:"true"` // true
	}

	ConfigRelMeAuthProvider struct {
		ID     string `env:"ID"`
		Secret string `env:"SECRET"`
		Type   string `env:"TYPE"`
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
