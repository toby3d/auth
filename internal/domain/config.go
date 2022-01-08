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
		StaticRootPath  string `yaml:"staticRootPath"`
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
		Length int           `yaml:"length"` // 32
	}

	ConfigJWT struct {
		Expiry      time.Duration `yaml:"expiry"` // 1h
		Secret      interface{}   `yaml:"secret"`
		Algorithm   string        `yaml:"algorithm"`   // HS256
		NonceLength int           `yaml:"nonceLength"` // 22
	}

	ConfigIndieAuth struct {
		Enabled bool `yaml:"enabled"` // true
	}

	ConfigTicketAuth struct {
		Expiry time.Duration `yaml:"expiry"` // 1m
		Length int           `yaml:"length"` // 24
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
			Secret:      []byte("hackme"),
			Algorithm:   "HS256",
		},
		IndieAuth: ConfigIndieAuth{
			Enabled: true,
		},
		TicketAuth: ConfigTicketAuth{
			Expiry: time.Minute,
			Length: 24,
		},
	}
}
