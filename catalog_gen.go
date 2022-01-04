// Code generated by running "go generate" in golang.org/x/text. DO NOT EDIT.

package main

import (
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/message/catalog"
)

type dictionary struct {
	index []uint32
	data  string
}

func (d *dictionary) Lookup(key string) (data string, ok bool) {
	p, ok := messageKeyToIndex[key]
	if !ok {
		return "", false
	}
	start, end := d.index[p], d.index[p+1]
	if start == end {
		return "", false
	}
	return d.data[start:end], true
}

func init() {
	dict := map[string]catalog.Dictionary{
		"en": &dictionary{index: enIndex, data: enData},
		"ru": &dictionary{index: ruIndex, data: ruData},
	}
	fallback := language.MustParse("en")
	cat, err := catalog.NewFromMap(dict, catalog.Fallback(fallback))
	if err != nil {
		panic(err)
	}
	message.DefaultCatalog = cat
}

var messageKeyToIndex = map[string]int{
	"Allow":   1,
	"Deny":    0,
	"Sign In": 2,
}

var enIndex = []uint32{ // 4 elements
	0x00000000, 0x00000005, 0x0000000b, 0x00000013,
} // Size: 40 bytes

const enData string = "\x02Deny\x02Allow\x02Sign In"

var ruIndex = []uint32{ // 4 elements
	0x00000000, 0x00000011, 0x00000024, 0x0000002f,
} // Size: 40 bytes

const ruData string = "" + // Size: 47 bytes
	"\x02Отказать\x02Разрешить\x02Войти"

	// Total table size 146 bytes (0KiB); checksum: 9261221B
