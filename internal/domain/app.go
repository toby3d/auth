package domain

type App struct {
	Name []string
	Logo []*URL
	URL  []*URL
}

// GetName safe returns first name, if any.
func (a App) GetName() string {
	if len(a.Name) == 0 {
		return ""
	}

	return a.Name[len(a.Name)-1]
}

// GetURL safe returns first uRL, if any.
func (a App) GetURL() *URL {
	if len(a.URL) == 0 {
		return nil
	}

	return a.URL[len(a.URL)-1]
}

// GetLogo safe returns first logo, if any.
func (a App) GetLogo() *URL {
	if len(a.Logo) == 0 {
		return nil
	}

	return a.Logo[len(a.Logo)-1]
}
