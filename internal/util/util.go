package util

import (
	"net"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	http "github.com/valyala/fasthttp"
	httputil "github.com/valyala/fasthttp/fasthttputil"
)

//nolint: exhaustivestruct
func TestServe(tb testing.TB, handler http.RequestHandler) (*http.Client, *http.Server, func()) {
	tb.Helper()

	ln := httputil.NewInmemoryListener()
	server := &http.Server{
		Handler:          handler,
		DisableKeepalive: true,
		CloseOnShutdown:  true,
	}

	go func() {
		assert.NoError(tb, server.Serve(ln))
	}()

	client := &http.Client{
		Dial: func(_ string) (net.Conn, error) {
			conn, err := ln.Dial()
			if err != nil {
				return nil, errors.Wrap(err, "cannot dial to address")
			}

			return conn, nil
		},
	}

	return client, server, func() {
		assert.NoError(tb, server.Shutdown())
	}
}
