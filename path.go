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
	obj := p.d.Get()
	var err error

	for _, i := range p.path {
		switch idx := i.(type) {
		case string:
			if obj.Kind() == KindVoid {
				return obj, nil
			}
			if obj.Kind() != KindMap {
				return nil, fmt.Errorf("%#v: tried to access property %#v of non-map %#v", p, idx, obj)
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
				return nil, fmt.Errorf("%#v: tried to access index %#v of non-list %#v", p, idx, obj)
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

func (p *Path) Put(v any) error {
	if len(p.path) == 0 {
		panic(fmt.Errorf("automerge: Path.Set called on path to Root()"))
	}

	_, set, err := p.ensure()
	if err != nil {
		return err
	}
	return set(v)
}

func (p *Path) Parent() *Path {
	if len(p.path) == 0 {
		panic(fmt.Errorf("automerge: Path.Parent called on path to Root()"))
	}
	return &Path{d: p.d, path: p.path[0 : len(p.path)-1]}
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

	return nil, fmt.Errorf("%#v: tried to write to key %v of non-map %#v", p, debugKey, v)
}

func (p *Path) ensureList(debugKey int) (*List, error) {
	if len(p.path) == 0 {
		return nil, fmt.Errorf("%#v: tried to write index %#v of non-list %#v", p, debugKey, p.d.Get())
	}

	v, set, err := p.ensure()
	if err != nil {
		return nil, err
	}

	if v.Kind() == KindVoid {
		t := NewList()
		if err := set(t); err != nil {
			return nil, err
		}
		return t, nil
	}

	if v.Kind() == KindList {
		return v.List(), nil
	}

	return nil, fmt.Errorf("%#v: tried to write to index %v of non-list %#v", p, p.path[len(p.path)-1], v)
}

func (p *Path) ensureText() (*Text, error) {
	if len(p.path) == 0 {
		return nil, fmt.Errorf("%#v: tried to edit non-text %#v", p, p.d.Get())
	}

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
	if len(p.path) == 0 {
		return nil, fmt.Errorf("%#v: tried to increment non-counter %#v", p, p.d.Get())
	}

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
	last := p.path[len(p.path)-1]
	switch key := last.(type) {
	case string:
		m, err := p.Parent().ensureMap(key)
		if err != nil {
			return nil, nil, err
		}
		v, err := m.Get(key)
		return v, func(v any) error {
			return m.Set(key, v)
		}, err

	case int:
		l, err := p.Parent().ensureList(key)
		if err != nil {
			return nil, nil, err
		}

		if key > l.Len() {
			return nil, nil, fmt.Errorf("%#p: index %v out of bounds (list.Len() == %v", p, key, l.Len())
		}
		v, err := l.Get(key)
		return v, func(v any) error {
			if key == l.Len() {
				return l.Append(v)
			}
			return l.Set(key, v)
		}, err

	default:
		panic("unreachable")
	}
}

// Map assumes there is a map at the given path.
// Calling methods on the map will error if the path cannot be traversed
// or if the value at this path is not a map.
// If there is a void at this location, writing to this map
// will implicitly create it (and the path as necessary).
func (p *Path) Map() *Map {
	return &Map{doc: p.d, path: p}
}

// List assumes there is a list at the given path.
// Calling methods on the list will error if the path cannot be traversed
// or if the value at this path is not a list.
// If there is a void at this location, reading from the list will return void,
// and writing to the list will implicitly create it (and the path as necesasry).
func (p *Path) List() *List {
	return &List{doc: p.d, path: p}
}

// Counter assumes there is a counter at the given path.
// Calling methods on the counter will error if the path cannot be traversed
// or if the value at this path is not a counter.
func (p *Path) Counter() *Counter {
	return &Counter{path: p}
}

// Counter assumes there is a counter at the given path.
// Calling methods on the counter will error if the path cannot be traversed
// or if the value at this path is not a counter.
func (p *Path) Text() *Text {
	return &Text{doc: p.d, path: p}
}

func (p *Path) GoString() string {
	str := "&automerge.Path("
	for i, p := range p.path {
		if i > 0 {
			str += ", "
		}
		str += fmt.Sprintf("%#v", p)
	}
	return str + ")"
}
