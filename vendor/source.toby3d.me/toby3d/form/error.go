package form

import (
	"reflect"
)

type (
	UnmarshalTypeError struct {
		Type   reflect.Type
		Value  string
		Struct string
		Field  string
		Offset int64
	}

	InvalidUnmarshalError struct {
		Type reflect.Type
	}
)

func (e UnmarshalTypeError) Error() string {
	if e.Struct != "" || e.Field != "" {
		return "form: cannot unmarshal " + e.Value + " into Go struct field " + e.Struct + "." +
			e.Field + " of type " + e.Type.String()
	}

	return "form: cannot unmarshal " + e.Value + " into Go value of type " + e.Type.String()
}

func (e InvalidUnmarshalError) Error() string {
	if e.Type == nil {
		return "form: Unmarshal(nil)"
	}

	if e.Type.Kind() != reflect.Pointer {
		return "form: Unmarshal(non-pointer " + e.Type.String() + "}"
	}

	return "form: Unmarshal(nil " + e.Type.String() + ")"
}
