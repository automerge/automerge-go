package automerge_test

import (
	"fmt"

	"github.com/ConradIrwin/automerge-go"
)

func ExampleAs() {
	doc := automerge.New(nil)
	doc.Path("isValid").Set(true)

	b, err := automerge.As[bool](doc.Path("isValid").Get())
	if err != nil {
		panic(err)
	}
	fmt.Println(b == true)

	type S struct {
		IsValid bool `json:"isValid"`
	}
	s, err := automerge.As[*S](doc.Root())
	if err != nil {
		panic(err)
	}
	fmt.Println(s.IsValid == true)
}

func ExampleList_Iter() {
	doc := automerge.New(nil)
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
	doc := automerge.New(nil)
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

func ExampleSyncState() {
	doc := automerge.New(nil)
	syncState := automerge.NewSyncState(doc)

	docUpdated := make(chan bool)
	recv := make(chan []byte)
	send := make(chan []byte)

loop:
	// generate an initial message, and then do so again
	// after receiving updates from the peer or making local changes
	for {
		msg, valid, err := syncState.GenerateMessage()
		if err != nil {
			panic(err)
		}
		if valid {
			send <- msg
		}

		select {
		case msg, ok := <-recv:
			if !ok {
				break loop
			}

			err := syncState.ReceiveMessage(msg)
			if err != nil {
				panic(err)
			}

		case _, ok := <-docUpdated:
			if !ok {
				break loop
			}
		}
	}
}
