// Code generated by qtc from "baseof.qtpl". DO NOT EDIT.
// See https://github.com/valyala/quicktemplate for details.

//line web/baseof.qtpl:1
package web

//line web/baseof.qtpl:1
import (
	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"source.toby3d.me/website/oauth/internal/domain"
)

//line web/baseof.qtpl:8
import (
	qtio422016 "io"

	qt422016 "github.com/valyala/quicktemplate"
)

//line web/baseof.qtpl:8
var (
	_ = qtio422016.Copy
	_ = qt422016.AcquireByteBuffer
)

//line web/baseof.qtpl:8
type Page interface {
//line web/baseof.qtpl:8
	Body() string
//line web/baseof.qtpl:8
	StreamBody(qw422016 *qt422016.Writer)
//line web/baseof.qtpl:8
	WriteBody(qq422016 qtio422016.Writer)
//line web/baseof.qtpl:8
	Head() string
//line web/baseof.qtpl:8
	StreamHead(qw422016 *qt422016.Writer)
//line web/baseof.qtpl:8
	WriteHead(qq422016 qtio422016.Writer)
//line web/baseof.qtpl:8
	Lang() string
//line web/baseof.qtpl:8
	StreamLang(qw422016 *qt422016.Writer)
//line web/baseof.qtpl:8
	WriteLang(qq422016 qtio422016.Writer)
//line web/baseof.qtpl:8
	Title() string
//line web/baseof.qtpl:8
	StreamTitle(qw422016 *qt422016.Writer)
//line web/baseof.qtpl:8
	WriteTitle(qq422016 qtio422016.Writer)
//line web/baseof.qtpl:8
}

//line web/baseof.qtpl:16
type BaseOf struct {
	Config   *domain.Config
	Language language.Tag
	Printer  *message.Printer
}

//line web/baseof.qtpl:24
func (p *BaseOf) StreamLang(qw422016 *qt422016.Writer) {
//line web/baseof.qtpl:25
	if p.Language != language.Und {
//line web/baseof.qtpl:26
		qw422016.E().S(p.Language.String())
//line web/baseof.qtpl:27
	} else {
//line web/baseof.qtpl:27
		qw422016.N().S(`en`)
//line web/baseof.qtpl:29
	}
//line web/baseof.qtpl:30
}

//line web/baseof.qtpl:30
func (p *BaseOf) WriteLang(qq422016 qtio422016.Writer) {
//line web/baseof.qtpl:30
	qw422016 := qt422016.AcquireWriter(qq422016)
//line web/baseof.qtpl:30
	p.StreamLang(qw422016)
//line web/baseof.qtpl:30
	qt422016.ReleaseWriter(qw422016)
//line web/baseof.qtpl:30
}

//line web/baseof.qtpl:30
func (p *BaseOf) Lang() string {
//line web/baseof.qtpl:30
	qb422016 := qt422016.AcquireByteBuffer()
//line web/baseof.qtpl:30
	p.WriteLang(qb422016)
//line web/baseof.qtpl:30
	qs422016 := string(qb422016.B)
//line web/baseof.qtpl:30
	qt422016.ReleaseByteBuffer(qb422016)
//line web/baseof.qtpl:30
	return qs422016
//line web/baseof.qtpl:30
}

//line web/baseof.qtpl:34
func (p *BaseOf) StreamTitle(qw422016 *qt422016.Writer) {
//line web/baseof.qtpl:34
	qw422016.N().S(` `)
//line web/baseof.qtpl:35
	qw422016.E().S(p.Config.Name)
//line web/baseof.qtpl:35
	qw422016.N().S(` `)
//line web/baseof.qtpl:36
}

//line web/baseof.qtpl:36
func (p *BaseOf) WriteTitle(qq422016 qtio422016.Writer) {
//line web/baseof.qtpl:36
	qw422016 := qt422016.AcquireWriter(qq422016)
//line web/baseof.qtpl:36
	p.StreamTitle(qw422016)
//line web/baseof.qtpl:36
	qt422016.ReleaseWriter(qw422016)
//line web/baseof.qtpl:36
}

//line web/baseof.qtpl:36
func (p *BaseOf) Title() string {
//line web/baseof.qtpl:36
	qb422016 := qt422016.AcquireByteBuffer()
//line web/baseof.qtpl:36
	p.WriteTitle(qb422016)
//line web/baseof.qtpl:36
	qs422016 := string(qb422016.B)
//line web/baseof.qtpl:36
	qt422016.ReleaseByteBuffer(qb422016)
//line web/baseof.qtpl:36
	return qs422016
//line web/baseof.qtpl:36
}

//line web/baseof.qtpl:38
func (p *BaseOf) StreamHead(qw422016 *qt422016.Writer) {
//line web/baseof.qtpl:38
	qw422016.N().S(` `)
//line web/baseof.qtpl:39
	qw422016.N().S(` <link rel="icon" href="`)
//line web/baseof.qtpl:42
	qw422016.E().S(p.Config.Server.StaticURLPrefix)
//line web/baseof.qtpl:42
	qw422016.N().S(`/favicon.ico" sizes="any"> <link rel="icon" href="`)
//line web/baseof.qtpl:47
	qw422016.E().S(p.Config.Server.StaticURLPrefix)
//line web/baseof.qtpl:47
	qw422016.N().S(`/icon.svg" type="image/svg+xml"> <link rel="apple-touch-icon" href="`)
//line web/baseof.qtpl:52
	qw422016.E().S(p.Config.Server.StaticURLPrefix)
//line web/baseof.qtpl:52
	qw422016.N().S(`/apple-touch-icon.png"> <link rel="manifest" href="`)
//line web/baseof.qtpl:56
	qw422016.E().S(p.Config.Server.StaticURLPrefix)
//line web/baseof.qtpl:56
	qw422016.N().S(`/manifest.webmanifest"> `)
//line web/baseof.qtpl:57
}

//line web/baseof.qtpl:57
func (p *BaseOf) WriteHead(qq422016 qtio422016.Writer) {
//line web/baseof.qtpl:57
	qw422016 := qt422016.AcquireWriter(qq422016)
//line web/baseof.qtpl:57
	p.StreamHead(qw422016)
//line web/baseof.qtpl:57
	qt422016.ReleaseWriter(qw422016)
//line web/baseof.qtpl:57
}

//line web/baseof.qtpl:57
func (p *BaseOf) Head() string {
//line web/baseof.qtpl:57
	qb422016 := qt422016.AcquireByteBuffer()
//line web/baseof.qtpl:57
	p.WriteHead(qb422016)
//line web/baseof.qtpl:57
	qs422016 := string(qb422016.B)
//line web/baseof.qtpl:57
	qt422016.ReleaseByteBuffer(qb422016)
//line web/baseof.qtpl:57
	return qs422016
//line web/baseof.qtpl:57
}

//line web/baseof.qtpl:59
func (p *BaseOf) StreamBody(qw422016 *qt422016.Writer) {
//line web/baseof.qtpl:59
}

//line web/baseof.qtpl:59
func (p *BaseOf) WriteBody(qq422016 qtio422016.Writer) {
//line web/baseof.qtpl:59
	qw422016 := qt422016.AcquireWriter(qq422016)
//line web/baseof.qtpl:59
	p.StreamBody(qw422016)
//line web/baseof.qtpl:59
	qt422016.ReleaseWriter(qw422016)
//line web/baseof.qtpl:59
}

//line web/baseof.qtpl:59
func (p *BaseOf) Body() string {
//line web/baseof.qtpl:59
	qb422016 := qt422016.AcquireByteBuffer()
//line web/baseof.qtpl:59
	p.WriteBody(qb422016)
//line web/baseof.qtpl:59
	qs422016 := string(qb422016.B)
//line web/baseof.qtpl:59
	qt422016.ReleaseByteBuffer(qb422016)
//line web/baseof.qtpl:59
	return qs422016
//line web/baseof.qtpl:59
}

//line web/baseof.qtpl:61
func StreamTemplate(qw422016 *qt422016.Writer, p Page) {
//line web/baseof.qtpl:61
	qw422016.N().S(` <!DOCTYPE html> <html class="page" lang="`)
//line web/baseof.qtpl:65
	p.StreamLang(qw422016)
//line web/baseof.qtpl:65
	qw422016.N().S(`"> <head> <meta charset="UTF-8"> <meta name="viewport" content="width=device-width, initial-scale=1.0"> `)
//line web/baseof.qtpl:73
	p.StreamHead(qw422016)
//line web/baseof.qtpl:73
	qw422016.N().S(` <title>`)
//line web/baseof.qtpl:74
	p.StreamTitle(qw422016)
//line web/baseof.qtpl:74
	qw422016.N().S(`</title> </head> <body class="page__body body"> `)
//line web/baseof.qtpl:77
	p.StreamBody(qw422016)
//line web/baseof.qtpl:77
	qw422016.N().S(` </body> </html> `)
//line web/baseof.qtpl:80
}

//line web/baseof.qtpl:80
func WriteTemplate(qq422016 qtio422016.Writer, p Page) {
//line web/baseof.qtpl:80
	qw422016 := qt422016.AcquireWriter(qq422016)
//line web/baseof.qtpl:80
	StreamTemplate(qw422016, p)
//line web/baseof.qtpl:80
	qt422016.ReleaseWriter(qw422016)
//line web/baseof.qtpl:80
}

//line web/baseof.qtpl:80
func Template(p Page) string {
//line web/baseof.qtpl:80
	qb422016 := qt422016.AcquireByteBuffer()
//line web/baseof.qtpl:80
	WriteTemplate(qb422016, p)
//line web/baseof.qtpl:80
	qs422016 := string(qb422016.B)
//line web/baseof.qtpl:80
	qt422016.ReleaseByteBuffer(qb422016)
//line web/baseof.qtpl:80
	return qs422016
//line web/baseof.qtpl:80
}

//line web/baseof.qtpl:82
func (p *BaseOf) StreamT(qw422016 *qt422016.Writer, str string, args ...interface{}) {
//line web/baseof.qtpl:82
	qw422016.N().S(` `)
//line web/baseof.qtpl:84
	result := p.Printer.Sprintf(str, args...)

//line web/baseof.qtpl:85
	qw422016.N().S(` `)
//line web/baseof.qtpl:86
	qw422016.E().S(result)
//line web/baseof.qtpl:86
	qw422016.N().S(` `)
//line web/baseof.qtpl:87
}

//line web/baseof.qtpl:87
func (p *BaseOf) WriteT(qq422016 qtio422016.Writer, str string, args ...interface{}) {
//line web/baseof.qtpl:87
	qw422016 := qt422016.AcquireWriter(qq422016)
//line web/baseof.qtpl:87
	p.StreamT(qw422016, str, args...)
//line web/baseof.qtpl:87
	qt422016.ReleaseWriter(qw422016)
//line web/baseof.qtpl:87
}

//line web/baseof.qtpl:87
func (p *BaseOf) T(str string, args ...interface{}) string {
//line web/baseof.qtpl:87
	qb422016 := qt422016.AcquireByteBuffer()
//line web/baseof.qtpl:87
	p.WriteT(qb422016, str, args...)
//line web/baseof.qtpl:87
	qs422016 := string(qb422016.B)
//line web/baseof.qtpl:87
	qt422016.ReleaseByteBuffer(qb422016)
//line web/baseof.qtpl:87
	return qs422016
//line web/baseof.qtpl:87
}
