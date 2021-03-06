// Code generated by qtc from "home.qtpl". DO NOT EDIT.
// See https://github.com/valyala/quicktemplate for details.

//line web/home.qtpl:1
package web

//line web/home.qtpl:1
import "source.toby3d.me/toby3d/auth/internal/domain"

//line web/home.qtpl:3
import (
	qtio422016 "io"

	qt422016 "github.com/valyala/quicktemplate"
)

//line web/home.qtpl:3
var (
	_ = qtio422016.Copy
	_ = qt422016.AcquireByteBuffer
)

//line web/home.qtpl:3
type HomePage struct {
	BaseOf
	Client *domain.Client
	State  string
}

//line web/home.qtpl:10
func (p *HomePage) StreamHead(qw422016 *qt422016.Writer) {
//line web/home.qtpl:10
	qw422016.N().S(` `)
//line web/home.qtpl:11
	p.BaseOf.StreamHead(qw422016)
//line web/home.qtpl:11
	qw422016.N().S(` `)
//line web/home.qtpl:12
	for i := range p.Client.RedirectURI {
//line web/home.qtpl:12
		qw422016.N().S(` <link rel="redirect_uri" href="`)
//line web/home.qtpl:13
		qw422016.E().S(p.Client.RedirectURI[i].String())
//line web/home.qtpl:13
		qw422016.N().S(`"> `)
//line web/home.qtpl:14
	}
//line web/home.qtpl:14
	qw422016.N().S(` `)
//line web/home.qtpl:15
}

//line web/home.qtpl:15
func (p *HomePage) WriteHead(qq422016 qtio422016.Writer) {
//line web/home.qtpl:15
	qw422016 := qt422016.AcquireWriter(qq422016)
//line web/home.qtpl:15
	p.StreamHead(qw422016)
//line web/home.qtpl:15
	qt422016.ReleaseWriter(qw422016)
//line web/home.qtpl:15
}

//line web/home.qtpl:15
func (p *HomePage) Head() string {
//line web/home.qtpl:15
	qb422016 := qt422016.AcquireByteBuffer()
//line web/home.qtpl:15
	p.WriteHead(qb422016)
//line web/home.qtpl:15
	qs422016 := string(qb422016.B)
//line web/home.qtpl:15
	qt422016.ReleaseByteBuffer(qb422016)
//line web/home.qtpl:15
	return qs422016
//line web/home.qtpl:15
}

//line web/home.qtpl:17
func (p *HomePage) StreamBody(qw422016 *qt422016.Writer) {
//line web/home.qtpl:17
	qw422016.N().S(` <header class="h-app h-x-app"> <img class="u-logo" src="`)
//line web/home.qtpl:21
	qw422016.E().S(p.Client.GetLogo().String())
//line web/home.qtpl:21
	qw422016.N().S(`" alt="`)
//line web/home.qtpl:22
	qw422016.E().S(p.Client.GetName())
//line web/home.qtpl:22
	qw422016.N().S(`" crossorigin="anonymous" decoding="async" height="140" importance="high" referrerpolicy="no-referrer-when-downgrade" width="140"> <h1> <a class="p-name u-url" href="`)
//line web/home.qtpl:33
	qw422016.E().S(p.Client.GetURL().String())
//line web/home.qtpl:33
	qw422016.N().S(`"> `)
//line web/home.qtpl:35
	qw422016.E().S(p.Client.GetName())
//line web/home.qtpl:35
	qw422016.N().S(` </a> </h1> </header> <main> <form method="get" action="/authorize" enctype="application/x-www-form-urlencoded" accept-charset="utf-8" target="_self"> `)
//line web/home.qtpl:48
	for name, value := range map[string]string{
		"client_id":     p.Client.ID.String(),
		"redirect_uri":  p.Client.RedirectURI[0].String(),
		"response_type": domain.ResponseTypeCode.String(),
		"state":         p.State,
		"scope":         domain.Scopes{domain.ScopeProfile, domain.ScopeEmail}.String(),
	} {
//line web/home.qtpl:54
		qw422016.N().S(` <input type="hidden" name="`)
//line web/home.qtpl:57
		qw422016.E().S(name)
//line web/home.qtpl:57
		qw422016.N().S(`" value="`)
//line web/home.qtpl:58
		qw422016.E().S(value)
//line web/home.qtpl:58
		qw422016.N().S(`"> `)
//line web/home.qtpl:59
	}
//line web/home.qtpl:59
	qw422016.N().S(` <input type="url" name="me" placeholder="https://example.com/" inputmode="url" autocomplete="url" required> <button type="submit">`)
//line web/home.qtpl:69
	p.StreamT(qw422016, "Sign In")
//line web/home.qtpl:69
	qw422016.N().S(`</button> </form> </main> `)
//line web/home.qtpl:72
}

//line web/home.qtpl:72
func (p *HomePage) WriteBody(qq422016 qtio422016.Writer) {
//line web/home.qtpl:72
	qw422016 := qt422016.AcquireWriter(qq422016)
//line web/home.qtpl:72
	p.StreamBody(qw422016)
//line web/home.qtpl:72
	qt422016.ReleaseWriter(qw422016)
//line web/home.qtpl:72
}

//line web/home.qtpl:72
func (p *HomePage) Body() string {
//line web/home.qtpl:72
	qb422016 := qt422016.AcquireByteBuffer()
//line web/home.qtpl:72
	p.WriteBody(qb422016)
//line web/home.qtpl:72
	qs422016 := string(qb422016.B)
//line web/home.qtpl:72
	qt422016.ReleaseByteBuffer(qb422016)
//line web/home.qtpl:72
	return qs422016
//line web/home.qtpl:72
}
