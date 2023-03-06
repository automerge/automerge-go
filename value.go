package automerge

import (
	"fmt"
	"strings"
	"time"
)

// Value represents a dynamically typed value read from a document.
// It can hold any of the supported primative types (bool, string, []byte, float64, int64, uint64, time.Time)
// the four mutable types (*Map, *List, *Text, *Counter), or it can be an explicit null,
// or a void to indicate that no value existed at all.
// You can convert from a Value to a go type using [As], or call accessor methods directly.
type Value struct {
	item *item
	doc  *Doc

	kind Kind
	val  any
}

func newValue(i *item, d *Doc) *Value {
	v := &Value{item: i, doc: d}
	v.kind = v.item.Kind()
	var objID *objID
	if v.kind == kindObjType {
		objID = v.item.objID()
		v.kind = objID.objKind(d)
	}

	switch v.kind {
	case KindNull, KindVoid, KindUnknown:
		// TODO: handle unknown values better?
		v.val = nil
	case KindBool:
		v.val = v.item.bool()
	case KindStr:
		v.val = v.item.str()
	case KindBytes:
		v.val = v.item.bytes()
	case KindFloat64:
		v.val = v.item.float64()
	case KindInt64:
		v.val = v.item.int64()
	case KindUint64:
		v.val = v.item.uint64()
	case KindTime:
		v.val = v.item.time()
	case KindCounter:
		v.val = v.item.counter()
	case KindMap:
		v.val = &Map{doc: d, objID: objID}
	case KindList:
		v.val = &List{doc: d, objID: objID}
	case KindText:
		v.val = &Text{doc: d, objID: objID}
	default:
		panic(fmt.Errorf("tried to create Value with Kind == %v", v.kind))
	}

	return v
}

func newValueInMap(i *item, m *Map, key string) *Value {
	v := newValue(i, m.doc)
	if c, ok := v.val.(*Counter); ok {
		c.m = m
		c.key = key
	}
	return v
}

func newValueInList(i *item, l *List, idx int) *Value {
	v := newValue(i, l.doc)
	if c, ok := v.val.(*Counter); ok {
		c.l = l
		c.idx = idx
	}
	return v
}

// Kind reports the kind of the value
func (v *Value) Kind() Kind {
	return v.kind
}

// IsVoid returns true if the value did not exist in the document
func (v *Value) IsVoid() bool {
	return v.Kind() == KindVoid
}

// IsNull returns true if the value is null
func (v *Value) IsNull() bool {
	return v.Kind() == KindNull
}

// IsUnknown returns true if the type of the value was unknown
func (v *Value) IsUnknown() bool {
	return v.kind == KindUnknown
}

// Bool returns the value as a bool, it panics if Kind() != KindBool
func (v *Value) Bool() bool {
	v.assertKind(KindBool)
	return v.val.(bool)
}

// Str returns the Value as a string, it panics if Kind() != KindStr
func (v *Value) Str() string {
	v.assertKind(KindStr)
	return v.val.(string)
}

// Bytes returns the Value as a []byte, it panics if Kind() != KindBytes
func (v *Value) Bytes() []byte {
	v.assertKind(KindBytes)
	return v.val.([]byte)
}

// Float64 returns the value as a float64, it panics if Kind() != KindFloat64
func (v *Value) Float64() float64 {
	v.assertKind(KindFloat64)
	return v.val.(float64)
}

// Int64 returns the value as a int64, it panics if Kind() != KindInt64
func (v *Value) Int64() int64 {
	v.assertKind(KindInt64)
	return v.val.(int64)
}

// Uint64 returns the value as a uint64, it panics if Kind() != KindUint64
func (v *Value) Uint64() uint64 {
	v.assertKind(KindUint64)
	return v.val.(uint64)
}

// Time returns the value as a time.Time, it panics if Kind() != KindTime
func (v *Value) Time() time.Time {
	v.assertKind(KindTime)
	return v.val.(time.Time)
}

// Map returns the value as a [*Map], it panics if Kind() != KindMap
func (v *Value) Map() *Map {
	v.assertKind(KindMap)
	return v.val.(*Map)
}

// List returns the value as a [*List], it panics if Kind() != KindList
func (v *Value) List() *List {
	v.assertKind(KindList)
	return v.val.(*List)
}

// Counter returns the value as a [*Counter], it panics if Kind() != KindCounter
func (v *Value) Counter() *Counter {
	v.assertKind(KindCounter)
	return v.val.(*Counter)
}

// Text returns the value as a [*Text], it panics if Kind() != KindText
func (v *Value) Text() *Text {
	v.assertKind(KindText)
	return v.val.(*Text)
}

func (v *Value) assertKind(k Kind) {
	if v.Kind() != k {
		panic(fmt.Errorf("automerge.Value: called .%s() on value of %v", strings.TrimPrefix(k.String(), "Kind"), v.Kind()))
	}
}

func (v *Value) goValue() any {
	switch v.kind {
	case KindMap:
		values, err := v.Map().Values()
		// this should not be able to happen because .load() is only
		// called from Value.goValue() which checks that this is a map.
		if err != nil {
			panic(err)
		}
		out := map[string]any{}
		for k, v := range values {
			out[k] = v.goValue()
		}
		return out
	case KindList:
		values, err := v.List().Values()
		// cannot happen as we know the value must have come from the doc
		if err != nil {
			panic(err)
		}
		out := []any{}
		for _, v := range values {
			out = append(out, v.goValue())
		}
		return out
	case KindText:
		s, err := v.Text().Get()
		// cannot happen as we know the value must have come from the doc
		if err != nil {
			panic(err)
		}
		return s
	case KindCounter:
		c, err := v.Counter().Get()
		// cannot happen as we know the value must have come from the doc
		if err != nil {
			panic(err)
		}
		return c
	default:
		return v.val
	}
}

// GoString returns a representation suitable for debugging.
func (v *Value) GoString() string {
	if v.kind == KindVoid {
		return "&automerge.Value(<void>)"
	}
	if v.kind == KindUnknown {
		return "&automerge.Value(<unknown>)"
	}
	return fmt.Sprintf("&automerge.Value(%#v)", v.val)
}

// String returns a representation suitable for debugging.
// Use [Value.Str] to get the underlying string.
func (v *Value) String() string {
	return v.GoString()
}
