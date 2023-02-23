package automerge

import (
	"encoding/json"
	"fmt"
	"time"
)

// normalize converts the value into a type expected by Put()
// bool/string/[]byte/int64/uint64/float64/time.Time/[]any/map[string]any/*Text/*Counter
func normalize(value any) (any, error) {
	if value == nil {
		return nil, nil
	}

	switch v := value.(type) {
	case bool, string, []byte, int64, uint64, float64,
		time.Time, map[string]any, []any,
		*Counter, *Text, *Map, *List:
		return value, nil
	case int:
		return float64(v), nil
	case int16:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case uint:
		return float64(v), nil
	case uint16:
		return float64(v), nil
	case uint32:
		return float64(v), nil
	}

	// TODO: this is a quite expensive approach...
	bytes, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("automerge: unsupported value: %#v", value)
	}

	if bytes[0] == '[' {
		value = []any{}
		if err := json.Unmarshal(bytes, &value); err != nil {
			return nil, fmt.Errorf("automerge: unsupported value: %#v", value)
		}
		return value, nil
	}
	if bytes[0] == '{' {
		value = map[string]any{}
		if err := json.Unmarshal(bytes, &value); err != nil {
			return nil, fmt.Errorf("automerge: unsupported value: %#v", value)
		}
		return value, nil
	}

	return nil, fmt.Errorf("automerge: unsupported value: %#v", value)
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

	switch (any)(ret).(type) {
	case *Map:
		if v.Kind() == KindMap {
			return v.val.(T), nil
		}
	case *List:
		if v.Kind() == KindList {
			return v.val.(T), nil
		}
	case *Counter:
		if v.Kind() == KindCounter {
			return v.val.(T), nil
		}
	case *Text:
		if v.Kind() == KindText {
			return v.val.(T), nil
		}

	default:
		var val any
		val = v.goValue()
		if r, ok := val.(T); ok {
			return r, nil
		}
		if bytes, err := json.Marshal(val); err == nil {
			if err == nil {
				if err := json.Unmarshal(bytes, &ret); err == nil {
					return ret, nil
				}
			}
		}
	}

	err = fmt.Errorf("automerge: could not convert %#v from %T to %T", v.val, v.val, ret)
	return
}
