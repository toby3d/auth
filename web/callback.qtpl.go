// Code generated by qtc from "callback.qtpl". DO NOT EDIT.
// See https://github.com/valyala/quicktemplate for details.

//line web/callback.qtpl:1
package web

//line web/callback.qtpl:1
import "source.toby3d.me/website/indieauth/internal/domain"

//line web/callback.qtpl:3
import (
	qtio422016 "io"

	qt422016 "github.com/valyala/quicktemplate"
)

//line web/callback.qtpl:3
var (
	_ = qtio422016.Copy
	_ = qt422016.AcquireByteBuffer
)

//line web/callback.qtpl:3
type CallbackPage struct {
	BaseOf
	Token *domain.Token
}

//line web/callback.qtpl:9
func (p *CallbackPage) StreamBody(qw422016 *qt422016.Writer) {
//line web/callback.qtpl:9
	qw422016.N().S(` `)
//line web/callback.qtpl:10
	if p.Token != nil {
//line web/callback.qtpl:10
		qw422016.N().S(` <h1>`)
//line web/callback.qtpl:11
		qw422016.E().S(p.Token.Me.String())
//line web/callback.qtpl:11
		qw422016.N().S(`</h1> <small>`)
//line web/callback.qtpl:12
		qw422016.E().S(p.Token.AccessToken)
//line web/callback.qtpl:12
		qw422016.N().S(`</small> `)
//line web/callback.qtpl:13
	}
//line web/callback.qtpl:13
	qw422016.N().S(` `)
//line web/callback.qtpl:14
}

//line web/callback.qtpl:14
func (p *CallbackPage) WriteBody(qq422016 qtio422016.Writer) {
//line web/callback.qtpl:14
	qw422016 := qt422016.AcquireWriter(qq422016)
//line web/callback.qtpl:14
	p.StreamBody(qw422016)
//line web/callback.qtpl:14
	qt422016.ReleaseWriter(qw422016)
//line web/callback.qtpl:14
}

//line web/callback.qtpl:14
func (p *CallbackPage) Body() string {
//line web/callback.qtpl:14
	qb422016 := qt422016.AcquireByteBuffer()
//line web/callback.qtpl:14
	p.WriteBody(qb422016)
//line web/callback.qtpl:14
	qs422016 := string(qb422016.B)
//line web/callback.qtpl:14
	qt422016.ReleaseByteBuffer(qb422016)
//line web/callback.qtpl:14
	return qs422016
//line web/callback.qtpl:14
}