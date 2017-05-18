package query

import (
	"fmt"
	"reflect"
)

type InvalidQueryError struct {
	Type reflect.Type
}

func (e *InvalidQueryError) Error() string {
	if e.Type == nil {
		return "invalid query: (nil)"
	}

	return "invalid query: (" + e.Type.String() + ")"
}

type DecodeTypeError struct {
	Type  reflect.Type
	Value string
}

func (e *DecodeTypeError) Error() string {
	return fmt.Sprintf("query: cannot decode %#v into field type %s", e.Value, e.Type.String())
}

type UnsupportedDecodeTypeError struct {
	Type reflect.Type
}

func (e *UnsupportedDecodeTypeError) Error() string {
	return fmt.Sprintf("query: unsupported decode type %s", e.Type.String())
}
