package model

type Client struct {
	ID          string
	URL         string
	Name        string
	Logo        string
	RedirectURI []string
}

func (Client) Bucket() []byte { return []byte("clients") }
