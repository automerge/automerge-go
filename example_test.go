package automerge_test

import (
	"fmt"

	"github.com/ConradIrwin/automerge-go"
)

func ExampleAs() {
	doc, _ := automerge.New(nil)
	doc.Path("isValid").Set(true)

	b, _ := automerge.As[bool](doc.Path("isValid").Get())
	fmt.Println(b == true)

	type S struct {
		IsValid bool `json:"isValid"`
	}
	s, _ := automerge.As[*S](doc.RootValue())
	fmt.Println(s.IsValid == true)
}

func ExampleList_Iter() {
	doc, _ := automerge.New(nil)
	list := doc.Path("list").List()

	iter := list.Iter()
	for {
		i, v, valid := iter.Next()
		if !valid {
			break
		}
		fmt.Println(i, v)
	}
	if iter.Error() != nil {
		panic(iter.Error())
	}
}

func ExampleMap_Iter() {
	doc, _ := automerge.New(nil)
	m := doc.Path("map").Map()

	iter := m.Iter()
	for {
		k, v, valid := iter.Next()
		if !valid {
			break
		}
		fmt.Println(k, v)
	}
	if iter.Error() != nil {
		panic(iter.Error())
	}
}
