package automerge

import (
	"fmt"
	"reflect"
)

type Struct struct {
	objID *objID
	doc   *Doc
}

func (d *Doc) SetRoot(v IsStruct) error {
	if reflect.ValueOf(v).IsNil() {
		return fmt.Errorf("cannot write nil to root")
	}
	s := v.isStruct()
	s.doc = d
	s.objID = rootObjID
	return s.write(v)
}

func (s *Struct) isStruct() *Struct {
	if s == nil {
		panic("tried to access nil *Struct: did you do struct{ *automerge.Struct } instead of struct{ automerge.Struct } by mistake?")
	}
	return s
}

func (s *Struct) asMap() *Map {
	return &Map{doc: s.doc, objID: s.objID}
}

func (s *Struct) write(v any) error {
	if s.objID == nil {
		panic("write called before objID")
	}

	m := s.asMap()
	mp, err := normalize(v, true)
	if err != nil {
		return err
	}
	repr, ok := mp.(map[string]any)
	if !ok {
		return fmt.Errorf("automerge: Struct.write called on non-struct: %#v %T %#v", v, v, mp)
	}
	for k, v := range repr {
		if err := m.Set(k, v); err != nil {
			return err
		}
	}
	return nil
}

type IsStruct interface {
	isStruct() *Struct
}

func NewStruct[T IsStruct](val T) T {
	return val
}
