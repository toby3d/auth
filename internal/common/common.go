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
	HeaderContentType string = "Content-Type"
)

const Und string = "und"
