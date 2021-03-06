// Automatically generated by internal/cmd/genreadfile/main.go. DO NOT EDIT

package jws

import (
	"io/fs"
	"os"
)

type sysFS struct{}

func (sysFS) Open(path string) (fs.File, error) {
	return os.Open(path)
}

func ReadFile(path string, options ...ReadFileOption) (*Message, error) {
	var parseOptions []ParseOption
	var readFileOptions []ReadFileOption
	for _, option := range options {
		if po, ok := option.(ParseOption); ok {
			parseOptions = append(parseOptions, po)
		} else {
			readFileOptions = append(readFileOptions, option)
		}
	}

	var srcFS fs.FS = sysFS{}
	for _, option := range options {
		switch option.Ident() {
		case identFS{}:
			srcFS = option.Value().(fs.FS)
		}
	}

	f, err := srcFS.Open(path)
	if err != nil {
		return nil, err
	}

	defer f.Close()
	return ParseReader(f)
}
