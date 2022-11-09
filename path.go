package automerge

import "fmt"

// Path is a cursor that lets you reach into the document
type Path struct {
	d    *Doc
	path []any
}

// Path extends the cursor with more path segments.
// It panics if any path segment is not a string or an int.
// It does not check that the path segment is traversable until you call
// a method that accesses the document.
func (p *Path) Path(path ...any) *Path {
	for _, v := range path {
		if _, ok := v.(string); !ok {
			if _, ok := v.(int); !ok {
				panic(fmt.Errorf("automerge: invalid path segment, expected string or int, got: %T(%#v)", v, v))
			}
		}
	}
	return &Path{d: p.d, path: append(p.path, path...)}
}

// Get returns the value at a given path
func (p *Path) Get() (*Value, error) {
	obj := p.d.RootValue()
	var err error

	for _, i := range p.path {
		switch idx := i.(type) {
		case string:
			if obj.Kind() == KindVoid {
				return obj, nil
			}
			if obj.Kind() != KindMap {
				return nil, fmt.Errorf("%#v: tried to read property %#v of non-map %#v", p, idx, obj.val)
			}
			obj, err = obj.Map().Get(idx)
			if err != nil {
				return nil, err
			}

		case int:
			if obj.Kind() == KindVoid {
				return obj, nil
			}
			if obj.Kind() != KindList {
				return nil, fmt.Errorf("%#v: tried to read index %#v of non-list %#v", p, idx, obj.val)
			}
			obj, err = obj.List().Get(idx)

			if err != nil {
				return nil, err
			}

		default:
			panic("unreachable")
		}
	}
	return obj, nil
}

// Set sets the value at the given path, and creates any missing parent
// Maps or Lists needed.
func (p *Path) Set(v any) error {
	_, set, err := p.ensure()
	if err != nil {
		return err
	}
	if err := set(v); err != nil {
		return err
	}
	return nil
}

func (p *Path) ensureMap(debugKey string) (*Map, error) {
	if len(p.path) == 0 {
		return p.d.Root(), nil
	}

	v, set, err := p.ensure()
	if err != nil {
		return nil, err
	}

	if v.Kind() == KindVoid {
		t := NewMap()
		if err := set(t); err != nil {
			return nil, err
		}
		return t, nil
	}

	if v.Kind() == KindMap {
		return v.Map(), nil
	}

	return nil, fmt.Errorf("%#v: tried to write property %#v of non-map %#v", p, debugKey, v.val)
}

func (p *Path) ensureList(debugKey int) (*List, error) {
	v, set, err := p.ensure()
	if err != nil {
		return nil, err
	}

	if v.Kind() == KindVoid {
		l := NewList()
		if err := set(l); err != nil {
			return nil, err
		}
		return l, nil
	}

	if v.Kind() == KindList {
		return v.List(), nil
	}

	return nil, fmt.Errorf("%#v: tried to write index %v of non-list %#v", p, debugKey, v.val)
}

func (p *Path) ensureText() (*Text, error) {
	v, set, err := p.ensure()
	if err != nil {
		return nil, err
	}

	if v.Kind() == KindVoid {
		t := NewText("")
		if err := set(t); err != nil {
			return nil, err
		}
		return t, nil
	}

	if v.Kind() == KindText {
		return v.Text(), nil
	}

	return nil, fmt.Errorf("%#v: tried to edit non-text %#v", p, v.val)
}

func (p *Path) ensureCounter() (*Counter, error) {
	v, set, err := p.ensure()
	if err != nil {
		return nil, err
	}

	if v.Kind() == KindVoid {
		t := NewCounter(0)
		if err := set(t); err != nil {
			return nil, err
		}
		return t, nil
	}

	if v.Kind() == KindCounter {
		return v.Counter(), nil
	}

	return nil, fmt.Errorf("%#v: tried to increment non-counter %#v", p, v.val)

}

func (p *Path) ensure() (*Value, func(v any) error, error) {
	if len(p.path) == 0 {
		return p.d.RootValue(), func(v any) error {
			return fmt.Errorf("%#v: tried to overwrite root of document", p)
		}, nil
	}
	last := p.path[len(p.path)-1]
	parent := &Path{d: p.d, path: p.path[0 : len(p.path)-1]}
	switch key := last.(type) {
	case string:
		m, err := parent.ensureMap(key)
		if err != nil {
			return nil, nil, err
		}
		v, err := m.Get(key)
		return v, func(v any) error {
			return m.Set(key, v)
		}, err

	case int:
		l, err := parent.ensureList(key)
		if err != nil {
			return nil, nil, err
		}

		v, err := l.Get(key)
		if err != nil {
			return nil, nil, err
		}
		return v, func(v any) error {
			if key > l.Len() {
				return fmt.Errorf("%#v: tried to write index %v beyond end of list length %v", p, key, l.Len())
			}
			if key == l.Len() {
				return l.Append(v)
			}
			return l.Set(key, v)
		}, nil

	default:
		panic("unreachable")
	}
}

// Map assumes there is a [Map] at the given path.
// Calling methods on the map will error if the path cannot be traversed
// or if the value at this path is not a map.
// If there is a void at this location, writing to this map
// will implicitly create it (and the path as necessary).
func (p *Path) Map() *Map {
	return &Map{doc: p.d, path: p}
}

// List assumes there is a [List] at the given path.
// Calling methods on the list will error if the path cannot be traversed
// or if the value at this path is not a list.
// If there is a void at this location, reading from the list will return void,
// and writing to the list will implicitly create it (and the path as necesasry).
func (p *Path) List() *List {
	return &List{doc: p.d, path: p}
}

// Counter assumes there is a [Counter] at the given path.
// Calling methods on the counter will error if the path cannot be traversed
// or if the value at this path is not a counter.
func (p *Path) Counter() *Counter {
	return &Counter{path: p}
}

// Text assumes there is a [Text] at the given path.
// Calling methods on the counter will error if the path cannot be traversed
// or if the value at this path is not a counter.
func (p *Path) Text() *Text {
	return &Text{doc: p.d, path: p}
}

// String returns a representation suitable for debugging.
func (p *Path) String() string {
	return p.GoString()
}

// GoString returns a representation suitable for debugging
func (p *Path) GoString() string {
	str := "&automerge.Path{"
	for i, p := range p.path {
		if i > 0 {
			str += ", "
		}
		str += fmt.Sprintf("%#v", p)
	}
	return str + "}"
}
