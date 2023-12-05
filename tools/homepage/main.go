// This package represent a simple tool for emulate simple localhost user domain
// onepage with support localhost IndieAuth instance.
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"source.toby3d.me/toby3d/auth/internal/common"
)

const (
	template string = `<!DOCKTYPE html>` +
		`<html>` +
		`<head>` +
		`<title>Playground homepage</title>` +
		`%[1]s` +
		`</head>` +
		`<body>` +
		`<form method="get" action="%[3]s/authorize">` +
		`<input type="url" name="me" placeholder="http://localhost/">` +
		`<input type="hidden" name="response_type" value="code">` +
		`<input type="hidden" name="state" value="hackme">` +
		`<input type="hidden" name="client_id" value="%[3]s/">` +
		`<input type="hidden" name="redirect_uri" value="%[3]s/callback">` +
		`<input type="hidden" name="scope" value="profile email">` +
		`<button type="submit">Sign In</button>` +
		`</form>` +
		`</body>` +
		`</html>`
	templateLinksDeprecated string = `<link rel="authorization_endpoint" href="%[1]s/authorize">` +
		`<link rel="token_endpoint" href="%[1]s/token">`
	templateLinksModern string = `<link rel="indieauth-metadata" href="%s/metadata">`
)

var (
	target     *string = flag.String("target", "http://127.0.0.1:3000", "set specific IndieAuth server addr")
	addr       *string = flag.String("addr", ":8080", "set specific port for pseudo homepage")
	deprecated *bool   = flag.Bool("deprecated", false, "enable deprecated handlers and links")
)

func main() {
	flag.Parse()

	log := log.New(os.Stdout, "homepage\t", log.LstdFlags)
	server := http.Server{
		Addr:                         *addr,
		Handler:                      NewHandler(*addr, *target, *deprecated),
		DisableGeneralOptionsHandler: false,
		TLSConfig:                    nil,
		ReadTimeout:                  500 * time.Millisecond,
		ReadHeaderTimeout:            0,
		WriteTimeout:                 500 * time.Millisecond,
		IdleTimeout:                  0,
		MaxHeaderBytes:               0,
		TLSNextProto:                 nil,
		ConnState:                    nil,
		ErrorLog:                     log,
		BaseContext:                  nil,
		ConnContext:                  nil,
	}

	log.Println("started pseudo homepage on", *addr)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalln(err)
	}
}

func NewHandler(addr, target string, deprecated bool) http.HandlerFunc {
	partial := fmt.Sprintf(templateLinksModern, target)

	if deprecated {
		partial = fmt.Sprintf(templateLinksDeprecated, target)
	}

	tpl := fmt.Sprintf(template, partial, addr, target)

	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set(common.HeaderContentType, common.MIMETextHTMLCharsetUTF8)
		fmt.Fprint(w, tpl)
	}
}
