// Package form implements encoding and decoding of urlencoded form. The mapping
// between form and Go values is described by `form:"query_name"` struct tags.
package form

import (
	"bytes"
	"fmt"
	"io"
	"net/url"
	"reflect"
	"strconv"
	"strings"
)

type (
	// Unmarshaler is the interface implemented by types that can unmarshal
	// a form description of themselves. The input can be assumed to be a
	// valid encoding of a form value. UnmarshalForm must copy the form data
	// if it wishes to retain the data after returning.
	Unmarshaler interface {
		UnmarshalForm(v []byte) error
	}

	Decoder struct {
		args url.Values
		tag  string
	}
)

const (
	tagIgnore    = "-"
	tagOmitempty = "omitempty"
	methodName   = "UnmarshalForm"
)

func NewDecoder(r io.Reader) *Decoder {
	buf := new(bytes.Buffer)
	defer buf.Reset()

	_, _ = buf.ReadFrom(r)
	args, _ := url.ParseQuery(buf.String())

	return &Decoder{
		tag:  "form",
		args: args,
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
func Unmarshal(data []byte, v any) error {
	return NewDecoder(bytes.NewReader(data)).Decode(v)
}

func (d Decoder) Decode(dst any) (err error) {
	src := reflect.ValueOf(dst)
	if !src.IsValid() || src.Kind() != reflect.Pointer || src.Elem().Kind() != reflect.Struct {
		return &InvalidUnmarshalError{
			Type: reflect.TypeOf(dst),
		}
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

	return d.decode("", src, "")
}

func (d Decoder) decode(key string, dst reflect.Value, opts tagOptions) error {
	src := d.args

	if keyIndex := strings.LastIndex(key, ","); keyIndex != -1 {
		if index, err := strconv.Atoi(key[keyIndex+1:]); err == nil {
			key = key[:keyIndex]

			src = make(url.Values)
			src.Set(key, d.args[key][index])
		}
	}

	switch dst.Kind() {
	case reflect.Bool:
		out, err := strconv.ParseBool(src.Get(key))
		if err != nil {
			return err
		}

		dst.SetBool(out)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		out, err := strconv.ParseInt(src.Get(key), 10, 64)
		if err != nil {
			return err
		}

		dst.SetInt(out)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		out, err := strconv.ParseUint(src.Get(key), 10, 64)
		if err != nil {
			return err
		}

		dst.SetUint(out)
	case reflect.Float32, reflect.Float64:
		out, err := strconv.ParseFloat(src.Get(key), 64)
		if err != nil {
			return err
		}

		dst.SetFloat(out)
	// case reflect.Array: // TODO(toby3d)
	// case reflect.Interface: // TODO(toby3d)
	case reflect.Slice:
		// NOTE(toby3d): copy raw []byte value as is
		if dst.Type().Elem().Kind() == reflect.Uint8 {
			dst.SetBytes([]byte(src.Get(key)))

			return nil
		}

		// NOTE(toby3d): if contains UnmarshalForm method
		for i := 0; i < dst.Addr().NumMethod(); i++ {
			if dst.Addr().Type().Method(i).Name != methodName {
				continue
			}

			in := make([]reflect.Value, 1)
			in[0] = reflect.ValueOf([]byte(src.Get(key)))

			out := dst.Addr().Method(i).Call(in)
			if len(out) > 0 && out[0].Interface() != nil && !opts.Contains(tagOmitempty) {
				return out[0].Interface().(error)
			}

			return nil
		}

		if dst.IsNil() {
			slice := d.args[key]
			dst.Set(reflect.MakeSlice(dst.Type(), len(slice), cap(slice)))
		}

		for i := 0; i < dst.Len(); i++ {
			if err := d.decode(fmt.Sprintf("%s,%d", key, i), dst.Index(i), ""); err != nil {
				return err
			}
		}
	case reflect.String:
		dst.SetString(string(src.Get(key)))
	case reflect.Pointer:
		if dst.IsNil() {
			dst.Set(reflect.New(dst.Type().Elem()))
		}

		// NOTE(toby3d): if contains UnmarshalForm method
		for i := 0; i < dst.NumMethod(); i++ {
			if dst.Type().Method(i).Name != methodName {
				continue
			}

			in := make([]reflect.Value, 1)
			in[0] = reflect.ValueOf([]byte(src.Get(key)))

			out := dst.Method(i).Call(in)
			if len(out) > 0 && out[0].Interface() != nil && !opts.Contains(tagOmitempty) {
				return out[0].Interface().(error)
			}

			return nil
		}

		if err := d.decode(key, dst.Elem(), ""); err != nil {
			return err
		}
	case reflect.Struct:
		// NOTE(toby3d): if contains UnmarshalForm method
		for i := 0; i < dst.Addr().NumMethod(); i++ {
			if dst.Addr().Type().Method(i).Name != methodName {
				continue
			}

			in := make([]reflect.Value, 1)
			in[0] = reflect.ValueOf([]byte(src.Get(key)))

			out := dst.Addr().Method(i).Call(in)
			if len(out) > 0 && out[0].Interface() != nil && !opts.Contains(tagOmitempty) {
				return out[0].Interface().(error)
			}

			return nil
		}

		for i := 0; i < dst.NumField(); i++ {
			if name, opts := parseTag(string(dst.Type().Field(i).Tag.Get(d.tag))); name != tagIgnore {
				if err := d.decode(name, dst.Field(i), opts); err != nil {
					return err
				}
			}
		}
	}

	return nil
}
