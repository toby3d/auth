package model

import (
	"fmt"

	"golang.org/x/xerrors"
)

//nolint: tagliatelle
type Error struct {
	Code        string        `json:"error"`
	Description string        `json:"error_description,omitempty"`
	URI         string        `json:"error_uri,omitempty"`
	Frame       xerrors.Frame `json:"-"`
}

func (e Error) Error() string {
	return fmt.Sprint(e)
}

func (e Error) Format(s fmt.State, r rune) {
	xerrors.FormatError(e, s, r)
}

func (e Error) FormatError(p xerrors.Printer) error {
	p.Printf("%s: %s", e.Code, e.Description)

	if e.URI != "" {
		p.Printf("%s", e.URI)
	}

	if p.Detail() {
		e.Frame.Format(p)
	}

	return nil
}
