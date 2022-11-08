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
	case bool, string, []byte, int64, uint64, float64, time.Time, map[string]any, []any, *Counter, *Text, *Map, *List, *void:
		return value, nil
	case int:
		return int64(v), nil
	case int32:
		return int64(v), nil
	case uint:
		return uint64(v), nil
	case uint32:
		return uint64(v), nil
	}

	// TODO: this is a quite expensive approach...
	bytes, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("unsupported value: %#v", value)
	}

	if bytes[0] == '[' {
		value = []any{}
		if err := json.Unmarshal(bytes, &value); err != nil {
			return nil, fmt.Errorf("unsupported value: %#v", value)
		}
		return value, nil
	}
	if bytes[0] == '{' {
		value = map[string]any{}
		if err := json.Unmarshal(bytes, &value); err != nil {
			return nil, fmt.Errorf("unsupported value: %#v", value)
		}
		return value, nil
	}

	return nil, fmt.Errorf("unsupported value: %#v", value)
}

func As[T any](v *Value, errs ...error) (ret T, err error) {
	if len(errs) > 0 && errs[0] != nil {
		err = errs[0]
		return
	}

	switch (interface{})(ret).(type) {
	case *Map:
		if v.Kind() == KindMap {
			return (interface{})(v.val).(T), nil
		}
	case *List:
		if v.Kind() == KindList {
			return (interface{})(v.val).(T), nil
		}
	case *Counter:
		if v.Kind() == KindCounter {
			return (interface{})(v.val).(T), nil
		}
	case *Text:
		if v.Kind() == KindText {
			return (interface{})(v.val).(T), nil
		}
	case *Value:
		return (interface{})(v).(T), nil

	default:
		var val any
		val, err = v.goValue()
		if err != nil {
			return
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
