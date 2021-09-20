package util

import (
	"net"

	http "github.com/valyala/fasthttp"
	httputil "github.com/valyala/fasthttp/fasthttputil"
)

func Serve(handler http.RequestHandler, req *http.Request, res *http.Response) error {
	ln := httputil.NewInmemoryListener()
	defer ln.Close()

	go func() {
		if err := http.Serve(ln, handler); err != nil {
			panic(err)
		}
	}()

	client := http.Client{
		Dial: func(addr string) (net.Conn, error) {
			return ln.Dial()
		},
	}

	return client.Do(req, res)
}
