//go:generate go run "$GOROOT/src/crypto/tls/generate_cert.go" --host 127.0.0.1,::1,localhost --start-date "Jan 1 00:00:00 1970" --duration=1000000h --ca --rsa-bits 1024 --ecdsa-curve P256
package httptest

import (
	"crypto/tls"

	// used for running tests.
	_ "embed"
	"net"
	"testing"
	"time"

	http "github.com/valyala/fasthttp"
	httputil "github.com/valyala/fasthttp/fasthttputil"
)

var (
	//go:embed cert.pem
	certData []byte
	//go:embed key.pem
	keyData []byte
)

// New returns the InMemory Server and the Client connected to it with the
// specified handler.
func New(tb testing.TB, handler http.RequestHandler) (*http.Client, *http.Server, func()) {
	tb.Helper()

	//nolint: exhaustivestruct
	server := &http.Server{
		CloseOnShutdown:   true,
		DisableKeepalive:  true,
		ReduceMemoryUsage: true,
		Handler:           http.TimeoutHandler(handler, 1*time.Second, "handler performance is too slow"),
	}

	ln := httputil.NewInmemoryListener()

	//nolint: errcheck
	go server.ServeTLSEmbed(ln, certData, keyData)

	//nolint: exhaustivestruct
	client := &http.Client{
		TLSConfig: &tls.Config{
			InsecureSkipVerify: true, //nolint: gosec
		},
		Dial: func(addr string) (net.Conn, error) {
			return ln.Dial() //nolint: wrapcheck
		},
	}

	return client, server, func() {
		_ = server.Shutdown()
	}
}

// NewRequest returns  a new incoming server Request and cleanup function.
func NewRequest(method, target string, body []byte) *http.Request {
	req := http.AcquireRequest()
	req.Header.SetMethod(method)
	req.SetRequestURI(target)
	req.SetBody(body)

	return req
}
