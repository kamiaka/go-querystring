package query

import (
	"net/url"
	"reflect"
	"strconv"
	"strings"
)

// Decode parses url.Values and stores the result in the value pointed to by v.
func Decode(values url.Values, v interface{}) error {
	rv := reflect.ValueOf(v)

	if rv.Kind() == reflect.Ptr && !rv.IsNil() {
		rv = rv.Elem()
	}

	if rv.Kind() != reflect.Struct {
		return &InvalidQueryError{reflect.TypeOf(v)}
	}

	_, err := decodeReflectStruct(values, rv, "")

	return err
}

func decodeReflectValue(values url.Values, rv reflect.Value, name string, opts tagOptions) (bool, error) {
	switch rv.Kind() {
	case reflect.Slice, reflect.Array:
		return decodeReflectArray(values, rv, name, opts)
	case reflect.Struct:
		return decodeReflectStruct(values, rv, name)
	case reflect.Interface:
		return decodeReflectValue(values, rv.Elem(), name, opts)
	case reflect.Ptr:
		orig := rv
		if rv.IsNil() {
			rv = reflect.New(rv.Type().Elem())
		}
		ok, err := decodeReflectValue(values, rv.Elem(), name, opts)
		if err != nil {
			return false, err
		}
		if ok && orig.IsNil() {
			orig.Set(rv)
		}
		return ok, nil
	case reflect.Invalid, reflect.Complex64, reflect.Complex128, reflect.Chan, reflect.Func, reflect.Map, reflect.UnsafePointer:
		return false, &UnsupportedDecodeTypeError{Type: rv.Type()}
	}

	vv, ok := values[name]
	if !ok {
		return false, nil
	}
	err := decodeReflectLiteral(vv[0], rv, opts)
	return ok, err
}

func decodeReflectStruct(values url.Values, rv reflect.Value, scope string) (bool, error) {
	rt := rv.Type()

	ok := false

	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)

		tag := field.Tag.Get("url")
		if tag == "" {
			tag = field.Name
		}
		tagName, tagOpts := parseTag(tag)

		name := tagName
		if scope != "" {
			name = scope + "[" + tagName + "]"
		}

		resp, err := decodeReflectValue(values, rv.Field(i), name, tagOpts)
		if err != nil {
			return false, err
		}
		ok = ok || resp
	}
	return ok, nil
}

func decodeReflectArray(values url.Values, rv reflect.Value, scope string, opts tagOptions) (bool, error) {
	name := scope

	sep := ""
	if opts.Contains("comma") {
		sep = ","
	} else if opts.Contains("space") {
		sep = " "
	} else if opts.Contains("semicolon") {
		sep = " "
	} else if opts.Contains("brackets") {
		name = name + "[]"
	}

	v, ok := values[name]
	if !ok {
		return false, nil
	}

	if sep != "" {
		v = strings.Split(v[0], sep)
	}

	return true, decodeReflectArrayLiteral(v, rv, opts)
}
func decodeReflectArrayLiteral(vv []string, rv reflect.Value, opts tagOptions) error {
	l := len(vv)
	if rv.Kind() == reflect.Slice {
		rv.Set(reflect.MakeSlice(rv.Type(), l, l))
	}
	for i := 0; i < rv.Len(); i++ {
		value := vv[i]
		err := decodeReflectLiteral(value, rv.Index(i), opts)
		if err != nil {
			return err
		}
	}
	return nil
}

func decodeReflectLiteral(value string, rv reflect.Value, opts tagOptions) error {
	rt := rv.Type()

	switch rt.Kind() {
	default:
		return &UnsupportedDecodeTypeError{Type: rv.Type()}
	case reflect.Bool:
		rv.SetBool(isTruthy(value))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v := int64(0)
		if value != "" {
			n, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return &DecodeTypeError{Type: rt, Value: value}
			}
			v = n
		}
		rv.SetInt(v)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v := uint64(0)
		if value != "" {
			n, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				return &DecodeTypeError{Type: rt, Value: value}
			}
			v = n
		}
		rv.SetUint(v)
	case reflect.Float32, reflect.Float64:
		v := float64(0)
		if value != "" {
			n, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return &DecodeTypeError{Type: rt, Value: value}
			}
			v = n
		}
		rv.SetFloat(v)
	case reflect.String:
		rv.SetString(value)
	}
	return nil
}

func isTruthy(s string) bool {
	switch s {
	case "0", "false", "FALSE", "False", "off", "OFF", "Off", "":
		return false
	}

	return true
}

type tagOptions []string

func parseTag(tag string) (string, tagOptions) {
	s := strings.Split(tag, ",")
	return s[0], s[1:]
}

func (b tagOptions) Contains(option string) bool {
	for _, s := range b {
		if s == option {
			return true
		}
	}
	return false
}
