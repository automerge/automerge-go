package automerge

import (
	"fmt"
	"runtime"
	"time"
)

// #include <automerge.h>
import "C"

// Map is an automerge type that stores a map of strings to values
type Map struct {
	doc *Doc

	objID *objID
	path  *Path
}

// NewMap returns a detached map.
// Before you can read from or write to it you must write it to the document.
func NewMap() *Map {
	return &Map{}
}

func (m *Map) lock() (*C.AMdoc, *C.AMobjId, func()) {
	cDoc, unlock := m.doc.lock()
	return cDoc, m.objID.cObjID, func() {
		runtime.KeepAlive(m)
		unlock()
	}
}

func (m *Map) createOnPath(key string) error {
	if m.path == nil {
		return nil
	}

	m2, err := m.path.ensureMap(key)
	if err != nil {
		return err
	}
	m.objID = m2.objID
	m.path = nil
	return nil
}

// Get retrieves the value from the map.
// This method will return an error if the underlying Get
// operation fails, or if this is the first attempt to access
// a Path.Map() and the path is not traverseable
func (m *Map) Get(key string) (*Value, error) {
	if m.doc == nil {
		return nil, fmt.Errorf("automerge.Map: tried to read detached map")
	}
	if m.path != nil {
		return m.path.Path(key).Get()
	}

	cKey, free := toByteSpanStr(key)
	defer free()
	cDoc, cObj, unlock := m.lock()
	defer unlock()

	item, err := wrap(C.AMmapGet(cDoc, cObj, cKey, nil)).item()
	if err != nil {
		return nil, err
	}
	return newValueInMap(item, m, key), nil
}

// Len returns the number of keys set in the map, or 0 on error
func (m *Map) Len() int {
	if m.doc == nil {
		return 0
	}
	if m.path != nil {
		v, err := m.path.Get()
		if err != nil || v.Kind() != KindMap {
			return 0
		}
		return v.Map().Len()
	}

	cDoc, cObj, unlock := m.lock()
	defer unlock()
	return int(C.AMobjSize(cDoc, cObj, nil))
}

// Del deletes a key and its corresponding value from the map
func (m *Map) Del(key string) error {
	if m.doc == nil {
		return fmt.Errorf("automerge.Map: tried to write to detached map")
	}
	if err := m.createOnPath(key); err != nil {
		return err
	}

	cKey, free := toByteSpanStr(key)
	defer free()
	cDoc, cObj, unlock := m.lock()
	defer unlock()

	return wrap(C.AMmapDelete(cDoc, cObj, cKey)).void()
}

// Set sets a key in the map to a given value.
// This method may error if the underlying operation errors,
// the type you provide cannot be converted to an automerge type,
// or if this is the first write to a [Path.Map] and the path is not traverseable.
func (m *Map) Set(key string, value any) error {
	if m.doc == nil {
		return fmt.Errorf("automerge.Map: tried to write to detached map")
	}
	if err := m.createOnPath(key); err != nil {
		return err
	}

	value, err := normalize(value)
	if err != nil {
		return err
	}

	cKey, free := toByteSpanStr(key)
	defer free()
	cDoc, cObj, unlock := m.lock()
	defer unlock()

	switch v := value.(type) {
	case nil:
		err = wrap(C.AMmapPutNull(cDoc, cObj, cKey)).void()

	case bool:
		err = wrap(C.AMmapPutBool(cDoc, cObj, cKey, C.bool(v))).void()
	case string:
		vStr, free := toByteSpanStr(v)
		defer free()
		err = wrap(C.AMmapPutStr(cDoc, cObj, cKey, vStr)).void()

	case []byte:
		vBytes, free := toByteSpan(v)
		defer free()
		err = wrap(C.AMmapPutBytes(cDoc, cObj, cKey, vBytes)).void()

	case int64:
		err = wrap(C.AMmapPutInt(cDoc, cObj, cKey, C.int64_t(v))).void()

	case uint64:
		err = wrap(C.AMmapPutUint(cDoc, cObj, cKey, C.uint64_t(v))).void()

	case float64:
		err = wrap(C.AMmapPutF64(cDoc, cObj, cKey, C.double(v))).void()

	case time.Time:
		err = wrap(C.AMmapPutTimestamp(cDoc, cObj, cKey, C.int64_t(v.UnixMilli()))).void()

	case []any:
		unlock()

		nl := NewList()
		if err := m.Set(key, nl); err != nil {
			return err

		}
		return nl.Append(v...)

	case map[string]any:
		unlock()

		n := NewMap()
		if err := m.Set(key, n); err != nil {
			return err
		}

		for key, val := range v {
			if err := n.Set(key, val); err != nil {
				return err
			}
		}

	case *Map:
		if v.objID != nil {
			return fmt.Errorf("automerge.Map: tried to move an existing *automerge.Map")
		}

		item, err := wrap(C.AMmapPutObject(cDoc, cObj, cKey, C.AM_OBJ_TYPE_MAP)).item()
		if err != nil {
			return err
		}

		v.doc = m.doc
		v.objID = item.objID()

	case *List:
		if v.objID != nil {
			return fmt.Errorf("automerge.Map: tried to move an existing *automerge.List")
		}

		item, err := wrap(C.AMmapPutObject(cDoc, cObj, cKey, C.AM_OBJ_TYPE_LIST)).item()
		if err != nil {
			return err
		}
		v.doc = m.doc
		v.objID = item.objID()

	case *Counter:
		if v.m != nil || v.l != nil {
			return fmt.Errorf("automerge.Map: tried to move an existing *automerge.Counter")
		}

		err = wrap(C.AMmapPutCounter(cDoc, cObj, cKey, C.int64_t(v.val))).void()
		if err == nil {
			v.m = m
			v.key = key
		}

	case *Text:
		if v.objID != nil {
			return fmt.Errorf("automerge.Map: tried to move an existing *automerge.Text")
		}
		item, err := wrap(C.AMmapPutObject(cDoc, cObj, cKey, C.AM_OBJ_TYPE_TEXT)).item()
		if err != nil {
			return err
		}
		v.doc = m.doc
		v.objID = item.objID()
		unlock()
		if err = v.Set(v.val); err != nil {
			return err
		}

	default:
		err = fmt.Errorf("automerge.Map: tried to write unsupported value %#v", value)
	}

	return err
}

func (m *Map) inc(key string, delta int64) error {
	cDoc, cObj, unlock := m.lock()
	defer unlock()
	cKey, free := toByteSpanStr(key)
	defer free()

	return wrap(C.AMmapIncrement(cDoc, cObj, cKey, C.int64_t(delta))).void()
}

// Values returns the values of the map
func (m *Map) Values() (map[string]*Value, error) {
	if m.doc == nil {
		return nil, fmt.Errorf("automerge.Map: tried to read detached map")
	}
	if m.path != nil {
		v, err := m.path.Get()
		if err != nil {
			return nil, err
		}
		switch v.Kind() {
		case KindMap:
			return v.Map().Values()
		case KindVoid:
			return nil, nil
		default:
			return nil, fmt.Errorf("%#v: tried to read non-map %#v", m.path, v.val)
		}
	}

	cDoc, cObj, unlock := m.lock()
	defer unlock()

	items, err := wrap(C.AMmapRange(cDoc, cObj, C.AMstr(nil), C.AMstr(nil), nil)).items()
	if err != nil {
		return nil, err
	}

	ret := map[string]*Value{}
	for _, i := range items {
		key := i.mapKey()
		ret[key] = newValueInMap(i, m, key)
	}
	return ret, nil
}

// Keys returns the current list of keys for the map
func (m *Map) Keys() ([]string, error) {
	v, err := m.Values()
	if err != nil {
		return nil, err
	}
	keys := []string{}
	for k := range v {
		keys = append(keys, k)
	}
	return keys, nil
}

// GoString returns a representation suitable for debugging.
func (m *Map) GoString() string {
	if m.doc == nil {
		return "&automerge.Map{}"
	}
	values, err := m.Values()
	if err != nil {
		return "&automerge.Map{<error>}"
	}

	sofar := "&automerge.Map{"
	i := 0
	for k, v := range values {
		if i > 0 {
			sofar += ", "
		}
		sofar += fmt.Sprintf("%#v: ", k)
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
		i++
	}

	return sofar + "}"
}

// String returns a representation suitable for debugging.
func (m *Map) String() string {
	return m.GoString()
}
