package common

const charsetUTF8 = "charset=UTF-8"

const (
	MIMEApplicationForm            string = "application/x-www-form-urlencoded"
	MIMEApplicationJSON            string = "application/json"
	MIMEApplicationJSONCharsetUTF8 string = MIMEApplicationJSON + "; " + charsetUTF8
	MIMETextHTML                   string = "text/html"
	MIMETextHTMLCharsetUTF8        string = MIMETextHTML + "; " + charsetUTF8
	MIMETextPlain                  string = "text/plain"
	MIMETextPlainCharsetUTF8       string = MIMETextPlain + "; " + charsetUTF8
)

const (
	HeaderAccept                   string = "Accept"
	HeaderAcceptLanguage           string = "Accept-Language"
	HeaderAccessControlAllowOrigin string = "Access-Control-Allow-Origin"
	HeaderAuthorization            string = "Authorization"
	HeaderContentType              string = "Content-Type"
	HeaderCookie                   string = "Cookie"
	HeaderHost                     string = "Host"
	HeaderLink                     string = "Link"
	HeaderLocation                 string = "Location"
	HeaderVary                     string = "Vary"
	HeaderWWWAuthenticate          string = "WWW-Authenticate"
	HeaderXCSRFToken               string = "X-CSRF-Token"
)

const Und string = "und"
