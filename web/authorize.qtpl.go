// Code generated by qtc from "authorize.qtpl". DO NOT EDIT.
// See https://github.com/valyala/quicktemplate for details.

//line web/authorize.qtpl:1
package web

//line web/authorize.qtpl:1
import "source.toby3d.me/toby3d/auth/internal/domain"

//line web/authorize.qtpl:3
import (
	qtio422016 "io"

	qt422016 "github.com/valyala/quicktemplate"
)

//line web/authorize.qtpl:3
var (
	_ = qtio422016.Copy
	_ = qt422016.AcquireByteBuffer
)

//line web/authorize.qtpl:3
type AuthorizePage struct {
	BaseOf
	CSRF                []byte
	Providers           []*domain.Provider
	Scope               domain.Scopes
	Client              *domain.Client
	Me                  *domain.Me
	RedirectURI         *domain.URL
	CodeChallengeMethod domain.CodeChallengeMethod
	ResponseType        domain.ResponseType
	CodeChallenge       string
	State               string
}

//line web/authorize.qtpl:17
func (p *AuthorizePage) StreamTitle(qw422016 *qt422016.Writer) {
//line web/authorize.qtpl:17
	qw422016.N().S(`
  `)
//line web/authorize.qtpl:18
	if p.Client.GetName() == "" {
//line web/authorize.qtpl:18
		qw422016.N().S(`
    `)
//line web/authorize.qtpl:19
		p.StreamT(qw422016, "Authorize %s", p.Client.GetName())
//line web/authorize.qtpl:19
		qw422016.N().S(`
  `)
//line web/authorize.qtpl:20
	} else {
//line web/authorize.qtpl:20
		qw422016.N().S(`
    `)
//line web/authorize.qtpl:21
		p.StreamT(qw422016, "Authorize application")
//line web/authorize.qtpl:21
		qw422016.N().S(`
  `)
//line web/authorize.qtpl:22
	}
//line web/authorize.qtpl:22
	qw422016.N().S(`
`)
//line web/authorize.qtpl:23
}

//line web/authorize.qtpl:23
func (p *AuthorizePage) WriteTitle(qq422016 qtio422016.Writer) {
//line web/authorize.qtpl:23
	qw422016 := qt422016.AcquireWriter(qq422016)
//line web/authorize.qtpl:23
	p.StreamTitle(qw422016)
//line web/authorize.qtpl:23
	qt422016.ReleaseWriter(qw422016)
//line web/authorize.qtpl:23
}

//line web/authorize.qtpl:23
func (p *AuthorizePage) Title() string {
//line web/authorize.qtpl:23
	qb422016 := qt422016.AcquireByteBuffer()
//line web/authorize.qtpl:23
	p.WriteTitle(qb422016)
//line web/authorize.qtpl:23
	qs422016 := string(qb422016.B)
//line web/authorize.qtpl:23
	qt422016.ReleaseByteBuffer(qb422016)
//line web/authorize.qtpl:23
	return qs422016
//line web/authorize.qtpl:23
}

//line web/authorize.qtpl:25
func (p *AuthorizePage) StreamBody(qw422016 *qt422016.Writer) {
//line web/authorize.qtpl:25
	qw422016.N().S(`
  <header>
    `)
//line web/authorize.qtpl:27
	if p.Client.GetLogo() != nil {
//line web/authorize.qtpl:27
		qw422016.N().S(`
      <img
        alt="`)
//line web/authorize.qtpl:29
		qw422016.E().S(p.Client.GetName())
//line web/authorize.qtpl:29
		qw422016.N().S(`"
        crossorigin="anonymous"
        decoding="async"
        height="140"
        importance="high"
        loading="lazy"
        referrerpolicy="no-referrer-when-downgrade"
        src="`)
//line web/authorize.qtpl:36
		qw422016.E().S(p.Client.GetLogo().String())
//line web/authorize.qtpl:36
		qw422016.N().S(`"
        width="140">
    `)
//line web/authorize.qtpl:38
	}
//line web/authorize.qtpl:38
	qw422016.N().S(`

    <h2>
      `)
//line web/authorize.qtpl:41
	if p.Client.GetURL() != nil {
//line web/authorize.qtpl:41
		qw422016.N().S(`
        <a href="`)
//line web/authorize.qtpl:42
		qw422016.E().S(p.Client.GetURL().String())
//line web/authorize.qtpl:42
		qw422016.N().S(`">
      `)
//line web/authorize.qtpl:43
	}
//line web/authorize.qtpl:43
	qw422016.N().S(`
      `)
//line web/authorize.qtpl:44
	if p.Client.GetName() != "" {
//line web/authorize.qtpl:44
		qw422016.N().S(`
        `)
//line web/authorize.qtpl:45
		qw422016.E().S(p.Client.GetName())
//line web/authorize.qtpl:45
		qw422016.N().S(`
      `)
//line web/authorize.qtpl:46
	} else {
//line web/authorize.qtpl:46
		qw422016.N().S(`
        `)
//line web/authorize.qtpl:47
		qw422016.E().S(p.Client.ID.String())
//line web/authorize.qtpl:47
		qw422016.N().S(`
      `)
//line web/authorize.qtpl:48
	}
//line web/authorize.qtpl:48
	qw422016.N().S(`
      `)
//line web/authorize.qtpl:49
	if p.Client.GetURL() != nil {
//line web/authorize.qtpl:49
		qw422016.N().S(`
        </a>
      `)
//line web/authorize.qtpl:51
	}
//line web/authorize.qtpl:51
	qw422016.N().S(`
    </h2>
  </header>

  <main>
    <form
      accept-charset="utf-8"
      action="/api/authorize"
      autocomplete="off"
      enctype="application/x-www-form-urlencoded"
      method="post"
      novalidate="true"
      target="_self">

      `)
//line web/authorize.qtpl:65
	if p.CSRF != nil {
//line web/authorize.qtpl:65
		qw422016.N().S(`
        <input
          type="hidden"
          name="_csrf"
          value="`)
//line web/authorize.qtpl:69
		qw422016.E().Z(p.CSRF)
//line web/authorize.qtpl:69
		qw422016.N().S(`">
      `)
//line web/authorize.qtpl:70
	}
//line web/authorize.qtpl:70
	qw422016.N().S(`

      `)
//line web/authorize.qtpl:72
	for key, val := range map[string]string{
		"client_id":     p.Client.ID.String(),
		"redirect_uri":  p.RedirectURI.String(),
		"response_type": p.ResponseType.String(),
		"state":         p.State,
	} {
//line web/authorize.qtpl:77
		qw422016.N().S(`
        <input
          type="hidden"
          name="`)
//line web/authorize.qtpl:80
		qw422016.E().S(key)
//line web/authorize.qtpl:80
		qw422016.N().S(`"
          value="`)
//line web/authorize.qtpl:81
		qw422016.E().S(val)
//line web/authorize.qtpl:81
		qw422016.N().S(`">
      `)
//line web/authorize.qtpl:82
	}
//line web/authorize.qtpl:82
	qw422016.N().S(`

      `)
//line web/authorize.qtpl:84
	if len(p.Scope) > 0 {
//line web/authorize.qtpl:84
		qw422016.N().S(`
      <fieldset>
        <legend>`)
//line web/authorize.qtpl:86
		p.StreamT(qw422016, "Choose your scopes")
//line web/authorize.qtpl:86
		qw422016.N().S(`</legend>

        `)
//line web/authorize.qtpl:88
		for _, scope := range p.Scope {
//line web/authorize.qtpl:88
			qw422016.N().S(`
          <div>
            <label>
              <input
                type="checkbox"
                name="scope[]"
                value="`)
//line web/authorize.qtpl:94
			qw422016.E().S(scope.String())
//line web/authorize.qtpl:94
			qw422016.N().S(`"
                checked>

              `)
//line web/authorize.qtpl:97
			qw422016.E().S(scope.String())
//line web/authorize.qtpl:97
			qw422016.N().S(`
            </label>
          </div>
        `)
//line web/authorize.qtpl:100
		}
//line web/authorize.qtpl:100
		qw422016.N().S(`
      </fieldset>
      `)
//line web/authorize.qtpl:102
	}
//line web/authorize.qtpl:102
	qw422016.N().S(`

      `)
//line web/authorize.qtpl:104
	if p.CodeChallenge != "" {
//line web/authorize.qtpl:104
		qw422016.N().S(`
        <input
          type="hidden"
          name="code_challenge"
          value="`)
//line web/authorize.qtpl:108
		qw422016.E().S(p.CodeChallenge)
//line web/authorize.qtpl:108
		qw422016.N().S(`">

        <input
          type="hidden"
          name="code_challenge_method"
          value="`)
//line web/authorize.qtpl:113
		qw422016.E().S(p.CodeChallengeMethod.String())
//line web/authorize.qtpl:113
		qw422016.N().S(`">
      `)
//line web/authorize.qtpl:114
	}
//line web/authorize.qtpl:114
	qw422016.N().S(`

      `)
//line web/authorize.qtpl:116
	if p.Me != nil {
//line web/authorize.qtpl:116
		qw422016.N().S(`
        <input
          type="hidden"
          name="me"
          value="`)
//line web/authorize.qtpl:120
		qw422016.E().S(p.Me.String())
//line web/authorize.qtpl:120
		qw422016.N().S(`">
      `)
//line web/authorize.qtpl:121
	}
//line web/authorize.qtpl:121
	qw422016.N().S(`

      `)
//line web/authorize.qtpl:123
	if len(p.Providers) > 0 {
//line web/authorize.qtpl:123
		qw422016.N().S(`
        <select
          name="provider"
          autocomplete
          required>

          `)
//line web/authorize.qtpl:129
		for _, provider := range p.Providers {
//line web/authorize.qtpl:129
			qw422016.N().S(`
            <option
              value="`)
//line web/authorize.qtpl:131
			qw422016.E().S(provider.UID)
//line web/authorize.qtpl:131
			qw422016.N().S(`"
              `)
//line web/authorize.qtpl:132
			if provider.UID == "mastodon" {
//line web/authorize.qtpl:132
				qw422016.N().S(`selected`)
//line web/authorize.qtpl:132
			}
//line web/authorize.qtpl:132
			qw422016.N().S(`>

              `)
//line web/authorize.qtpl:134
			qw422016.E().S(provider.Name)
//line web/authorize.qtpl:134
			qw422016.N().S(`
            </option>
          `)
//line web/authorize.qtpl:136
		}
//line web/authorize.qtpl:136
		qw422016.N().S(`
        </select>
      `)
//line web/authorize.qtpl:138
	} else {
//line web/authorize.qtpl:138
		qw422016.N().S(`
        <input type="hidden" name="provider" value="direct">
      `)
//line web/authorize.qtpl:140
	}
//line web/authorize.qtpl:140
	qw422016.N().S(`

      <button
        type="submit"
        name="authorize"
        value="deny">

        `)
//line web/authorize.qtpl:147
	p.StreamT(qw422016, "Deny")
//line web/authorize.qtpl:147
	qw422016.N().S(`
      </button>

      <button
        type="submit"
        name="authorize"
        value="allow">

        `)
//line web/authorize.qtpl:155
	p.StreamT(qw422016, "Allow")
//line web/authorize.qtpl:155
	qw422016.N().S(`
      </button>
    </form>
  </main>
`)
//line web/authorize.qtpl:159
}

//line web/authorize.qtpl:159
func (p *AuthorizePage) WriteBody(qq422016 qtio422016.Writer) {
//line web/authorize.qtpl:159
	qw422016 := qt422016.AcquireWriter(qq422016)
//line web/authorize.qtpl:159
	p.StreamBody(qw422016)
//line web/authorize.qtpl:159
	qt422016.ReleaseWriter(qw422016)
//line web/authorize.qtpl:159
}

//line web/authorize.qtpl:159
func (p *AuthorizePage) Body() string {
//line web/authorize.qtpl:159
	qb422016 := qt422016.AcquireByteBuffer()
//line web/authorize.qtpl:159
	p.WriteBody(qb422016)
//line web/authorize.qtpl:159
	qs422016 := string(qb422016.B)
//line web/authorize.qtpl:159
	qt422016.ReleaseByteBuffer(qb422016)
//line web/authorize.qtpl:159
	return qs422016
//line web/authorize.qtpl:159
}
