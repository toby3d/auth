package model

type Profile struct {
	Name  string
	URL   URL
	Photo URL
	Email string
}

func NewProfile() *Profile {
	return new(Profile)
}
