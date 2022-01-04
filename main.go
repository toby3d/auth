//go:generate go install github.com/valyala/quicktemplate/qtc@latest
//go:generate qtc -dir=./web
//go:generate go install golang.org/x/text/cmd/gotext@latest
//go:generate gotext -srclang=en update -out=catalog_gen.go -lang=en,ru
package main

import (
	_ "embed"

	"source.toby3d.me/website/indieauth/internal/cmd"
)

func main() { cmd.Execute() }
