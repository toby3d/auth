//go:generate go run "$GOROOT/src/crypto/tls/generate_cert.go" --host 127.0.0.1,::1,localhost --start-date "Jan 1 00:00:00 1970" --duration=1000000h --ca --rsa-bits 1024 --ecdsa-curve P256
package util

import (
	"crypto/tls"
	"net"
	"path/filepath"
	"testing"
	"time"

	http "github.com/valyala/fasthttp"
	httputil "github.com/valyala/fasthttp/fasthttputil"
)

//nolint: exhaustivestruct
func TestServe(tb testing.TB, handler http.RequestHandler) (*http.Client, *http.Server, func()) {
	tb.Helper()

	server := &http.Server{
		CloseOnShutdown:   true,
		DisableKeepalive:  true,
		Handler:           http.TimeoutHandler(handler, 1*time.Minute, "handler performance is too slow"),
		ReduceMemoryUsage: true,
	}

	ln := httputil.NewInmemoryListener()

	//nolint: errcheck
	go server.ServeTLS(ln, filepath.Join("..", "..", "..", "util", "cert.pem"),
		filepath.Join("..", "..", "..", "util", "key.pem"))

	client := &http.Client{
		TLSConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
		Dial: func(addr string) (net.Conn, error) {
			return ln.Dial()
		},
	}

	return client, server, func() {
		server.Shutdown() //nolint: errcheck
	}
}
