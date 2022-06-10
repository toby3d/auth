package form

import (
	"strings"
)

type tagOptions string

const delim rune = ','

func parseTag(tag string) (string, tagOptions) {
	tag, opt, _ := strings.Cut(tag, string(delim))

	return tag, tagOptions(opt)
}

func (o tagOptions) Contains(optionName string) bool {
	if len(o) == 0 {
		return false
	}

	s := string(o)
	for s != "" {
		var name string
		if name, s, _ = strings.Cut(s, string(delim)); name == optionName {
			return true
		}
	}

	return false
}
