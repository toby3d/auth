package model

type Client struct {
	ID          URL
	Name        URL
	Logo        URL
	URL         URL
	RedirectURI []URL
}

func NewClient() *Client {
	c := new(Client)
	c.RedirectURI = make([]URL, 0)

	return c
}
