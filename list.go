package automerge

import (
	"fmt"
	"runtime"
	"time"
)

// #include "automerge.h"
import "C"

// List is an automerge type that stores a list of [Value]'s
type List struct {
	doc   *Doc
	objID *objID
	path  *Path
}

// NewList returns a detached list.
// Before you can read from or write to it you must write it to the document.
func NewList() *List {
	return &List{}
}

func (l *List) lock() (*C.AMdoc, *C.AMobjId, func()) {
	cDoc, unlock := l.doc.lock()
	return cDoc, l.objID.cObjID, func() {
		runtime.KeepAlive(l)
		unlock()
	}
}

// Len returns the length of the list, or 0 on error
func (l *List) Len() int {
	if l.doc == nil {
		return 0
	}
	if l.path != nil {
		v, err := l.path.Get()
		if err != nil || v.Kind() != KindList {
			return 0
		}
		return v.List().Len()
	}

	cDoc, cObj, unlock := l.lock()
	defer unlock()
	return int(C.AMobjSize(cDoc, cObj, nil))
}

// Values returns a slice of the values in a list
func (l *List) Values() ([]*Value, error) {
	if l.doc == nil {
		return nil, fmt.Errorf("automerge.List: tried to read detached list")
	}
	if l.path != nil {
		v, err := l.path.Get()
		if err != nil {
			return nil, err
		}
		switch v.Kind() {
		case KindList:
			return v.List().Values()
		case KindVoid:
			return nil, nil
		default:
			return nil, fmt.Errorf("%#v: tried to read non-list %#v", l.path, v.val)
		}
	}
	cDoc, cObj, unlock := l.lock()
	defer unlock()

	items, err := wrap(C.AMlistRange(cDoc, cObj, 0, C.SIZE_MAX, nil)).items()
	if err != nil {
		return nil, err
	}

	ret := []*Value{}
	for i, item := range items {
		ret = append(ret, newValueInList(item, l, i))
	}
	return ret, nil
}

// Get returns the value at index i
func (l *List) Get(i int) (*Value, error) {
	if l.doc == nil {
		return nil, fmt.Errorf("automerge.List: tried to read detached list")
	}
	if l.path != nil {
		return l.path.Path(i).Get()
	}

	// make lists act more like maps
	if i < 0 || i >= l.Len() {
		return &Value{kind: KindVoid}, nil
	}

	cDoc, cObj, unlock := l.lock()
	defer unlock()

	item, err := wrap(C.AMlistGet(cDoc, cObj, C.size_t(i), nil)).item()
	if err != nil {
		return nil, err
	}
	return newValueInList(item, l, i), nil
}

// Append adds the values at the end of the list.
func (l *List) Append(values ...any) error {
	for _, v := range values {
		if err := l.put(C.SIZE_MAX, true, v); err != nil {
			return err
		}
	}
	return nil
}

// Set overwrites the value at l[idx] with value.
func (l *List) Set(idx int, value any) error {
	if idx < 0 || idx >= l.Len() {
		return fmt.Errorf("automerge.List: tried to write index %v beyond end of list length %v", idx, l.Len())
	}
	return l.put(C.size_t(idx), false, value)
}

// Insert inserts the new values just before idx.
func (l *List) Insert(idx int, value ...any) error {
	if idx < 0 || idx > l.Len() {
		return fmt.Errorf("automerge.List: tried to write index %v beyond end of list length %v", idx, l.Len())
	}
	for i, v := range value {
		if err := l.put(C.size_t(idx+i), true, v); err != nil {
			return err
		}
	}
	return nil
}

// Delete removes the value at idx and shortens the list.
func (l *List) Delete(idx int) error {
	if idx < 0 || idx >= l.Len() {
		return fmt.Errorf("automerge.List: tried to write index %v beyond end of list length %v", idx, l.Len())
	}

	cDoc, cObj, unlock := l.lock()
	defer unlock()

	return wrap(C.AMlistDelete(cDoc, cObj, C.size_t(idx))).void()
}

func (l *List) inc(i int, delta int64) error {
	cDoc, cObj, unlock := l.lock()
	defer unlock()

	return wrap(C.AMlistIncrement(cDoc, cObj, C.size_t(i), C.int64_t(delta))).void()
}

func (l *List) put(i C.size_t, before bool, value any) error {
	if l.doc == nil {
		return fmt.Errorf("automerge.List: tried to write to detached list")
	}
	if l.path != nil {
		l2, err := l.path.ensureList(int(i))
		if err != nil {
			return err
		}
		l.objID = l2.objID
		l.path = nil
	}

	value, err := normalize(value, false)
	if err != nil {
		return err
	}

	cDoc, cObj, unlock := l.lock()
	defer unlock()

	switch v := value.(type) {
	case nil:
		err = wrap(C.AMlistPutNull(cDoc, cObj, i, C.bool(before))).void()
	case bool:
		err = wrap(C.AMlistPutBool(cDoc, cObj, i, C.bool(before), C.bool(v))).void()
	case string:
		vStr, free := toByteSpanStr(v)
		defer free()
		err = wrap(C.AMlistPutStr(cDoc, cObj, i, C.bool(before), vStr)).void()

	case []byte:
		vBytes, free := toByteSpan(v)
		defer free()
		err = wrap(C.AMlistPutBytes(cDoc, cObj, i, C.bool(before), vBytes)).void()

	case int64:
		err = wrap(C.AMlistPutInt(cDoc, cObj, i, C.bool(before), C.int64_t(v))).void()

	case uint64:
		err = wrap(C.AMlistPutUint(cDoc, cObj, i, C.bool(before), C.uint64_t(v))).void()

	case float64:
		err = wrap(C.AMlistPutF64(cDoc, cObj, i, C.bool(before), C.double(v))).void()

	case time.Time:
		err = wrap(C.AMlistPutTimestamp(cDoc, cObj, i, C.bool(before), C.int64_t(v.UnixMilli()))).void()

	case []any:
		unlock()

		nl := NewList()
		if err := l.put(i, before, nl); err != nil {
			return err
		}
		return nl.Append(v...)

	case map[string]any:
		unlock()

		nm := NewMap()
		if err := l.put(i, before, nm); err != nil {
			return err
		}
		for key, val := range v {
			if err := nm.Set(key, val); err != nil {
				return err
			}
		}

	case *Counter:
		if v.m != nil || v.l != nil {
			return fmt.Errorf("automerge.List: tried to move an attached *automerge.Text")
		}

		err = wrap(C.AMlistPutCounter(cDoc, cObj, i, C.bool(before), C.int64_t(v.val))).void()
		if err == nil {
			v.l = l
			v.idx = int(i)
		}

	case *Text:
		if v.objID != nil {
			return fmt.Errorf("automerge.List: tried to move an attached *automerge.Text")
		}
		item, err := wrap(C.AMlistPutObject(cDoc, cObj, i, C.bool(before), C.AM_OBJ_TYPE_TEXT)).item()
		if err != nil {
			return err
		}
		v.doc = l.doc
		v.objID = item.objID()
		unlock()
		if err := v.Set(v.val); err != nil {
			return err
		}

	case *Map:
		if v.objID != nil {
			return fmt.Errorf("automerge.List: tried to move an attached *automerge.Map")
		}
		item, err := wrap(C.AMlistPutObject(cDoc, cObj, i, C.bool(before), C.AM_OBJ_TYPE_MAP)).item()
		if err != nil {
			return err
		}
		v.doc = l.doc
		v.objID = item.objID()

	case *List:
		if v.objID != nil {
			return fmt.Errorf("automerge.List: tried to move an attached *automerge.List")
		}
		item, err := wrap(C.AMlistPutObject(cDoc, cObj, i, C.bool(before), C.AM_OBJ_TYPE_LIST)).item()
		if err != nil {
			return err
		}
		v.doc = l.doc
		v.objID = item.objID()

	case IsStruct:
		s := v.isStruct()
		if s.objID != nil {
			return fmt.Errorf("automerge.List: tried to move an attached *automerge.Map")
		}
		item, err := wrap(C.AMlistPutObject(cDoc, cObj, i, C.bool(before), C.AM_OBJ_TYPE_MAP)).item()
		if err != nil {
			return err
		}
		s.doc = l.doc
		s.objID = item.objID()
		s.write(v)

	default:
		err = fmt.Errorf("automerge.List: tried to write unsupported value %#v", value)
	}

	return err
}

// GoString returns a representation suitable for debugging.
func (l *List) GoString() string {
	if l.doc == nil {
		return "&automerge.List{}"
	}
	values, err := l.Values()
	if err != nil {
		return "&automerge.List{<error>}"
	}
	sofar := "&automerge.List{"
	for i, v := range values {
		if i > 0 {
			sofar += ", "
		}
		i++
		if v.Kind() == KindMap {
			sofar += "&automerge.Map{...}"
		} else if v.Kind() == KindList {
			sofar += "&automerge.List{...}"
		} else {
			sofar += fmt.Sprintf("%#v", v.val)
		}

		if i >= 5 {
			sofar += ", ..."
			break
		}
	}

	return sofar + "}"
}

// String returns a representation suitable for debugging.
func (l *List) String() string {
	return l.GoString()
}
