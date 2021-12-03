// Package form implements encoding and decoding of urlencoded form. The mapping
// between form and Go values is described by `form:"query_name"` struct tags.
package form

import (
	"errors"
	"reflect"

	http "github.com/valyala/fasthttp"
)

type (
	// Unmarshaler is the interface implemented by types that can unmarshal
	// a form description of themselves. The input can be assumed to be a
	// valid encoding of a form value. UnmarshalForm must copy the form data
	// if it wishes to retain the data after returning.
	//
	// By convention, to approximate the behavior of Unmarshal itself,
	// Unmarshalers implement UnmarshalForm([]byte("null")) as a no-op.
	Unmarshaler interface {
		UnmarshalForm(v []byte) error
	}

	// A Decoder reads and decodes form values from an *fasthttp.Args.
	Decoder struct {
		source *http.Args
	}
)

const tagName string = "form"

// NewDecoder returns a new decoder that reads from *fasthttp.Args.
func NewDecoder(args *http.Args) *Decoder {
	return &Decoder{
		source: args,
	}
}

// Decode reads the next form-encoded value from its input and stores it in the
// value pointed to by v.
func (dec *Decoder) Decode(v interface{}) error {
	dst := reflect.ValueOf(v).Elem()
	if !dst.IsValid() {
		return errors.New("invalid input")
	}

	st := reflect.TypeOf(v).Elem()

	for i := 0; i < dst.NumField(); i++ {
		field := st.Field(i)

		// NOTE(toby3d): get tag value as query name
		tagValue, ok := field.Tag.Lookup(tagName)
		if !ok || tagValue == "" || tagValue == "-" || !dec.source.Has(tagValue) {
			continue
		}

		// NOTE(toby3d): read struct field type
		switch field.Type.Kind() {
		case reflect.String:
			dst.Field(i).SetString(string(dec.source.Peek(tagValue)))
		case reflect.Int:
			dst.Field(i).SetInt(int64(dec.source.GetUintOrZero(tagValue)))
		case reflect.Float64:
			dst.Field(i).SetFloat(dec.source.GetUfloatOrZero(tagValue))
		case reflect.Bool:
			dst.Field(i).SetBool(dec.source.GetBool(tagValue))
		case reflect.Ptr: // NOTE(toby3d): pointer to another struct
			// NOTE(toby3d): check what custom unmarshal method exists
			beforeFunc := dst.Field(i).MethodByName("UnmarshalForm")
			if beforeFunc.IsNil() {
				continue
			}

			dst.Field(i).Set(reflect.New(field.Type.Elem()))
			beforeFunc.Call([]reflect.Value{reflect.ValueOf(dec.source.Peek(tagValue))})
		case reflect.Slice:
			switch field.Type.Elem().Kind() {
			case reflect.Uint8: // NOTE(toby3d): bytes slice
				dst.Field(i).SetBytes(dec.source.Peek(tagValue))
			case reflect.String: // NOTE(toby3d): string slice
				values := dec.source.PeekMulti(tagValue)
				slice := reflect.MakeSlice(field.Type, len(values), len(values))

				for j, v := range values {
					slice.Index(j).SetString(string(v))
				}

				dst.Field(i).Set(slice)
			}
		}
	}

	return nil
}
