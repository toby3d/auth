package domain

import (
	"path"
	"strings"

	http "github.com/valyala/fasthttp"
)

// Provider represent 3rd party RelMeAuth provider.
type Provider struct {
	Scopes       []string
	AuthURL      string
	ClientID     string
	ClientSecret string
	Name         string
	Photo        string
	RedirectURL  string
	TokenURL     string
	UID          string
	URL          string
}

//nolint:gochecknoglobals // structs cannot be contants
var (
	ProviderDirect = Provider{
		AuthURL:      "/authorize",
		ClientID:     "",
		ClientSecret: "",
		Name:         "IndieAuth",
		Photo:        path.Join("static", "icon.svg"),
		RedirectURL:  path.Join("callback"),
		Scopes:       []string{},
		TokenURL:     "/token",
		UID:          "direct",
		URL:          "/",
	}

	ProviderTwitter = Provider{
		AuthURL:      "https://twitter.com/i/oauth2/authorize",
		ClientID:     "",
		ClientSecret: "",
		Name:         "Twitter",
		Photo:        path.Join("static", "providers", "twitter.svg"),
		RedirectURL:  path.Join("callback", "twitter"),
		Scopes:       []string{"tweet.read", "users.read"},
		TokenURL:     "https://api.twitter.com/2/oauth2/token",
		UID:          "twitter",
		URL:          "https://twitter.com/",
	}

	ProviderGitHub = Provider{
		AuthURL:      "https://github.com/login/oauth/authorize",
		ClientID:     "",
		ClientSecret: "",
		Name:         "GitHub",
		Photo:        path.Join("static", "providers", "github.svg"),
		RedirectURL:  path.Join("callback", "github"),
		Scopes:       []string{"read:user", "user:email"},
		TokenURL:     "https://github.com/login/oauth/access_token",
		UID:          "github",
		URL:          "https://github.com/",
	}

	ProviderGitLab = Provider{
		AuthURL:      "https://gitlab.com/oauth/authorize",
		ClientID:     "",
		ClientSecret: "",
		Name:         "GitLab",
		Photo:        path.Join("static", "providers", "gitlab.svg"),
		RedirectURL:  path.Join("callback", "gitlab"),
		Scopes:       []string{"read_user"},
		TokenURL:     "https://gitlab.com/oauth/token",
		UID:          "gitlab",
		URL:          "https://gitlab.com/",
	}

	ProviderMastodon = Provider{
		AuthURL:      "https://mstdn.io/oauth/authorize",
		ClientID:     "",
		ClientSecret: "",
		Name:         "Mastodon",
		Photo:        path.Join("static", "providers", "mastodon.svg"),
		RedirectURL:  path.Join("callback", "mastodon"),
		Scopes:       []string{"read:accounts"},
		TokenURL:     "https://mstdn.io/oauth/token",
		UID:          "mastodon",
		URL:          "https://mstdn.io/",
	}
)

// AuthCodeURL returns URL for authorize user in RelMeAuth client.
func (p Provider) AuthCodeURL(state string) string {
	uri := http.AcquireURI()
	defer http.ReleaseURI(uri)
	uri.Update(p.AuthURL)

	for key, val := range map[string]string{
		"client_id":     p.ClientID,
		"redirect_uri":  p.RedirectURL,
		"response_type": "code",
		"scope":         strings.Join(p.Scopes, " "),
		"state":         state,
	} {
		uri.QueryArgs().Set(key, val)
	}

	return uri.String()
}
