package automerge

import (
	"fmt"
	"reflect"
	"strings"
	"time"
)

var (
	boolType   = reflect.TypeOf(true)
	stringType = reflect.TypeOf("")
)

// normalize converts the value into a type expected by Put()
// bool/string/[]byte/int64/uint64/float64/time.Time/[]any/map[string]any/*Text/*Counter
func normalize(value any, isStruct bool) (any, error) {
	if value == nil {
		return nil, nil
	}
	if _, ok := value.(IsStruct); ok && !isStruct {
		return value, nil
	}

	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.Bool:
		return rv.Bool(), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32:
		return float64(rv.Int()), nil
	case reflect.Int64:
		return rv.Int(), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32:
		return float64(rv.Uint()), nil
	case reflect.Uint64:
		return rv.Uint(), nil
	case reflect.Float32, reflect.Float64:
		return rv.Float(), nil
	case reflect.String:
		return rv.String(), nil
	case reflect.Interface:
		return normalize(rv.Elem().Interface(), false)

	case reflect.Pointer:
		switch reflect.TypeOf(value) {
		case reflect.TypeOf(&Map{}):
			return value.(*Map), nil
		case reflect.TypeOf(&Text{}):
			return value.(*Text), nil
		case reflect.TypeOf(&List{}):
			return value.(*List), nil
		case reflect.TypeOf(&Counter{}):
			return value.(*Counter), nil
		}
		return normalize(rv.Elem().Interface(), false)

	case reflect.Slice, reflect.Array:
		if rv.Type().Elem().Kind() == reflect.Uint8 {
			ret := []byte{}
			for i := 0; i < rv.Len(); i++ {
				ret = append(ret, byte(rv.Index(i).Uint()))
			}
			return ret, nil
		}

		ret := []any{}

		for i := 0; i < rv.Len(); i++ {
			v, err := normalize(rv.Index(i).Interface(), false)
			if err != nil {
				return nil, err
			}
			ret = append(ret, v)
		}
		return ret, nil

	case reflect.Map:
		ret := map[string]any{}

		if rv.Type().Key().Kind() != reflect.String {
			return nil, fmt.Errorf("automerge: unsupported map, must have string keys")
		}

		for _, k := range rv.MapKeys() {
			v, err := normalize(rv.MapIndex(k).Interface(), false)
			if err != nil {
				return nil, err
			}
			ret[k.String()] = v
		}
		return ret, nil

	case reflect.Struct:
		t := rv.Type()

		if t == reflect.TypeOf(time.Time{}) {
			return value, nil
		}

		ret := map[string]any{}

		for i := 0; i < t.NumField(); i++ {
			ft := t.Field(i)
			name, omit := parseTags(ft)
			if omit {
				continue
			}

			v, err := normalize(rv.Field(i).Interface(), false)
			if err != nil {
				return nil, err
			}
			ret[name] = v
		}

		return ret, nil

	// Invalid, Complex64, Complex128, Chan, Func, Uintptr, UnsafePointer
	default:
		return nil, fmt.Errorf("automerge: unsupported type %v", rv.Kind())
	}
}

func parseTags(ft reflect.StructField) (name string, omit bool) {
	tag := ft.Tag.Get("automerge")
	if tag == "-" || !ft.IsExported() {
		return "", true
	}
	name, _, _ = strings.Cut(tag, ",")
	if name == "" {
		name = ft.Name
	}
	return name, false
}

// As converts v to type T.
// If the v cannot be converted to T, an error will be returned.
// T can be any of the automerge builtin types ([*Map], [*List], [*Counter], [*Text]),
// or any type that v's underlying value could be converted to. This conversion
// works generously, so that (for example) T could be any numeric type that v's value
// fits in, or if v.IsNull() then the 0-value of T will be returned.
//
// To make it easier to wrap a call to Get(), when an not-nil error is passed
// as a second argument, As will return (nil, err)
func As[T any](v *Value, errs ...error) (ret T, err error) {
	if len(errs) > 0 && errs[0] != nil {
		err = errs[0]
		return
	}

	err = unmarshal(reflect.ValueOf(&ret).Elem(), v)
	return
}

func unmarshalNumber[T interface{ int64 | float64 | uint64 }](overflows func(T) bool, set func(T), v *Value) bool {
	switch v.Kind() {
	case KindFloat64:
		f := v.Float64()
		if float64(T(f)) == f && !overflows(T(f)) {
			set(T(f))
			return true
		}
	case KindInt64:
		i := v.Int64()
		if (T(i) > 0) == (i > 0) && !overflows(T(i)) {
			set(T(i))
			return true
		}
	case KindUint64:
		u := v.Uint64()
		if T(u) > 0 && !overflows(T(u)) {
			set(T(u))
			return true
		}
	case KindCounter:
		i := v.Counter().val
		if (T(i) > 0) == (i > 0) && !overflows(T(i)) {
			set(T(i))
			return true
		}
	}
	return false
}

func unmarshal(rv reflect.Value, v *Value) error {
	switch rv.Kind() {
	case reflect.Bool:
		if v.Kind() == KindBool {
			rv.SetBool(v.Bool())
			return nil
		}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if unmarshalNumber(rv.OverflowInt, rv.SetInt, v) {
			return nil
		}

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if unmarshalNumber(rv.OverflowUint, rv.SetUint, v) {
			return nil
		}

	case reflect.Float32, reflect.Float64:
		if unmarshalNumber(rv.OverflowFloat, rv.SetFloat, v) {
			return nil
		}

	case reflect.String:
		switch v.Kind() {
		case KindStr:
			rv.SetString(v.Str())
			return nil
		case KindText:
			s, err := v.Text().Get()
			if err != nil {
				return err
			}
			rv.SetString(s)
			return nil
		}

	case reflect.Interface:
		val := reflect.ValueOf(v.Interface())
		if !val.IsValid() {
			// reflect.ValueOf(nil) returns the zero Value which cannot be used.
			// reflect.Zero() returns the Value representing the nil value of the current type,
			// which can be...
			val = reflect.Zero(rv.Type())
		}
		rv.Set(val)
		return nil

	case reflect.Ptr:
		if rv.Type() == reflect.TypeOf(&Map{}) && v.Kind() == KindMap {
			rv.Set(reflect.ValueOf(v.val))
			return nil
		}
		if rv.Type() == reflect.TypeOf(&List{}) && v.Kind() == KindList {
			rv.Set(reflect.ValueOf(v.val))
			return nil
		}
		if rv.Type() == reflect.TypeOf(&Counter{}) && v.Kind() == KindCounter {
			rv.Set(reflect.ValueOf(v.val))
			return nil
		}
		if rv.Type() == reflect.TypeOf(&Text{}) && v.Kind() == KindText {
			rv.Set(reflect.ValueOf(v.val))
			return nil
		}

		r := reflect.New(rv.Type().Elem())
		if err := unmarshal(r.Elem(), v); err != nil {
			return err
		}
		rv.Set(r)
		return nil

	case reflect.Slice:
		if v.Kind() == KindBytes && rv.Type().AssignableTo(reflect.TypeOf([]byte{})) {
			rv.Set(reflect.ValueOf(v.Bytes()))
			return nil
		}
		if v.Kind() != KindList {
			break
		}

		vals, err := v.List().Values()
		if err != nil {
			return err
		}

		nl := reflect.New(rv.Type()).Elem()
		for _, lv := range vals {
			v := reflect.New(rv.Type().Elem())
			if err := unmarshal(v.Elem(), lv); err != nil {
				return err
			}
			nl = reflect.Append(nl, v.Elem())
		}
		rv.Set(nl)
		return nil

	case reflect.Array:
		if v.Kind() != KindList {
			break
		}

		vals, err := v.List().Values()
		if err != nil {
			return err
		}
		max := len(vals)
		if rv.Len() < max {
			max = rv.Len()
		}

		for i := 0; i < max; i++ {
			if err := unmarshal(rv.Index(i), vals[i]); err != nil {
				return err
			}
		}
		return nil

	case reflect.Map:
		if v.Kind() != KindMap {
			return fmt.Errorf("automerge: cannot unmarshal %s into %s", v.Kind(), rv.Type().String())
		}

		if rv.Type().Key().Kind() != reflect.String {
			return fmt.Errorf("automerge: unsupported map, must have string keys")
		}

		vals, err := v.Map().Values()
		if err != nil {
			return err
		}

		nm := reflect.MakeMapWithSize(rv.Type(), len(vals))

		for mk, mv := range vals {
			rk := reflect.New(rv.Type().Key())
			rk.Elem().SetString(mk)

			rv := reflect.New(rv.Type().Elem())
			if err := unmarshal(rv.Elem(), mv); err != nil {
				return err
			}

			nm.SetMapIndex(rk.Elem(), rv.Elem())
		}

		rv.Set(nm)
		return nil

	case reflect.Struct:
		if v.Kind() == KindTime && rv.Type().AssignableTo(reflect.TypeOf(time.Now())) {
			rv.Set(reflect.ValueOf(v.Time()))
			return nil
		}

		if v.Kind() != KindMap {
			return fmt.Errorf("automerge: cannot unmarshal %s into %s", v.Kind(), rv.Type().String())
		}

		vals, err := v.Map().Values()
		if err != nil {
			return err
		}

		s := reflect.New(rv.Type()).Elem()

		for i := 0; i < rv.Type().NumField(); i++ {
			ft := rv.Type().Field(i)
			name, omit := parseTags(ft)
			if omit {
				continue
			}

			if mv, ok := vals[name]; ok {
				if err := unmarshal(s.Field(i), mv); err != nil {
					return err
				}
			}

			rv.Set(s)
		}
		return nil

	default:
		return fmt.Errorf("automerge: unsupported type %v", rv.Kind())
	}

	return fmt.Errorf("automerge: cannot unmarshal %s into %s", v.Kind(), rv.Type().String())
}
