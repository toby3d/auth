// Package form implements encoding and decoding of urlencoded form. The mapping
// between form and Go values is described by `form:"query_name"` struct tags.
package form

import (
	"errors"
	"fmt"
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

// Unmarshal parses the form-encoded data and stores the result in the value
// pointed to by v. If v is nil or not a pointer, Unmarshal returns error.
//
// Unmarshal uses the reflection, allocating maps, slices, and pointers as
// necessary, with the following additional rules:
//
// To unmarshal form into a pointer, Unmarshal first handles the case of the
// form being the form literal null. In that case, Unmarshal sets the pointer to
// nil. Otherwise, Unmarshal unmarshals the form into the value pointed at by
// the pointer. If the pointer is nil, Unmarshal allocates a new value for it to
// point to.
//
// To unmarshal form into a value implementing the Unmarshaler interface,
// Unmarshal calls that value's UnmarshalForm method, including when the input
// is a form null.
//
// To unmarshal form into a struct, Unmarshal matches incoming object keys to
// the keys (either the struct field name or its tag), preferring an exact match
// but also accepting a case-insensitive match. By default, object keys which
// don't have a corresponding struct field are ignored.
func Unmarshal(src *http.Args, dst interface{}) error {
	if err := NewDecoder(src).Decode(dst); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	return nil
}

// Decode reads the next form-encoded value from its input and stores it in the
// value pointed to by v.
func (dec *Decoder) Decode(src interface{}) (err error) {
	v := reflect.ValueOf(src).Elem()
	if !v.IsValid() {
		return errors.New("invalid input")
	}

	defer func() {
		if r := recover(); r != nil {
			if ve, ok := r.(*reflect.ValueError); ok {
				err = fmt.Errorf("recovered: %w", ve)
			} else {
				panic(r)
			}
		}
	}()

	t := reflect.TypeOf(src).Elem()

	for i := 0; i < v.NumField(); i++ {
		ft := t.Field(i)

		// NOTE(toby3d): get tag value as query name
		tagValue, ok := ft.Tag.Lookup(tagName)
		if !ok || tagValue == "" || tagValue == "-" || !dec.source.Has(tagValue) {
			continue
		}

		field := v.Field(i)

		// NOTE(toby3d): read struct field type
		switch ft.Type.Kind() {
		case reflect.String:
			field.SetString(string(dec.source.Peek(tagValue)))
		case reflect.Int:
			field.SetInt(int64(dec.source.GetUintOrZero(tagValue)))
		case reflect.Float64:
			field.SetFloat(dec.source.GetUfloatOrZero(tagValue))
		case reflect.Bool:
			field.SetBool(dec.source.GetBool(tagValue))
		case reflect.Ptr: // NOTE(toby3d): pointer to another struct
			field.Set(reflect.New(ft.Type.Elem()))

			// NOTE(toby3d): check what custom unmarshal method exists
			unmarshalFunc := field.MethodByName("UnmarshalForm")
			if unmarshalFunc.IsZero() {
				continue
			}

			unmarshalFunc.Call([]reflect.Value{reflect.ValueOf(dec.source.Peek(tagValue))})
		case reflect.Slice:
			switch ft.Type.Elem().Kind() {
			case reflect.Uint8: // NOTE(toby3d): bytes slice
				field.SetBytes(dec.source.Peek(tagValue))
			case reflect.String: // NOTE(toby3d): string slice
				values := dec.source.PeekMulti(tagValue)
				slice := reflect.MakeSlice(ft.Type, len(values), len(values))

				for j, vv := range values {
					slice.Index(j).SetString(string(vv))
				}

				field.Set(slice)
			}
		}
	}

	return
}
