package domain

import "net/url"

type App struct {
	Logo []*url.URL
	URL  []*url.URL
	Name []string
}

// GetName safe returns first name, if any.
func (a App) GetName() string {
	if len(a.Name) == 0 {
		return ""
	}

	return a.Name[0]
}

// GetURL safe returns first URL, if any.
func (a App) GetURL() *url.URL {
	if len(a.URL) == 0 {
		return nil
	}

	return a.URL[0]
}

// GetLogo safe returns first logo, if any.
func (a App) GetLogo() *url.URL {
	if len(a.Logo) == 0 {
		return nil
	}

	return a.Logo[0]
}
