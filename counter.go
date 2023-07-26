package automerge

import "fmt"

// Counter is a mutable int64 that collaborators
// can increment or decrement.
type Counter struct {
	val int64

	path *Path

	m   *Map
	key string
	l   *List
	idx int
}

// NewCounter returns a detached counter with the given starting value.
// Before you can call [Counter.Get] or [Counter.Inc] you must write it to the document.
func NewCounter(v int64) *Counter {
	return &Counter{val: v}
}

// Get retrurns the current value of the counter.
func (c *Counter) Get() (int64, error) {
	var v *Value
	var err error
	if c.m != nil {
		v, err = c.m.Get(c.key)
	} else if c.l != nil {
		v, err = c.l.Get(c.idx)
	} else if c.path != nil {
		v, err = c.path.Get()
	} else {
		return 0, fmt.Errorf("automerge.Counter: tried to read from detached counter")
	}
	if err != nil {
		return 0, err
	}
	switch v.Kind() {
	case KindVoid:
		return 0, nil
	case KindCounter:
		return v.Counter().val, nil
	default:
		return 0, fmt.Errorf("automerge.Counter: tried to read non-counter %#v", v.val)
	}
}

// Inc adjusts the counter by delta.
func (c *Counter) Inc(delta int64) error {
	if c.m == nil && c.l == nil {
		if c.path == nil {
			return fmt.Errorf("automerge.Counter: tried to write to detached counter")
		}

		c2, err := c.path.ensureCounter()
		if err != nil {
			return err
		}
		c.l = c2.l
		c.idx = c2.idx
		c.m = c2.m
		c.key = c2.key
		c.path = nil
	}

	if c.l != nil {
		return c.l.inc(c.idx, delta)
	}
	return c.m.inc(c.key, delta)
}

// GoString returns a representation suitable for debugging.
func (c *Counter) GoString() string {
	if c.l == nil && c.m == nil && c.path == nil {
		return fmt.Sprintf("&automerge.Counter{%v}", c.val)
	}
	v, err := c.Get()
	if err != nil {
		return "&automerge.Counter{<error>}"
	}
	return fmt.Sprintf("&automerge.Counter{%v}", v)
}

// String returns a representation suitable for debugging.
func (c *Counter) String() string {
	return c.GoString()
}
