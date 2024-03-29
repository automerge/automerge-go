package automerge_test

import (
	"fmt"

	"github.com/automerge/automerge-go"
)

func ExampleAs() {
	doc := automerge.New()
	doc.Path("isValid").Set(true)
	doc.Path("foo", "bar").Set("baz")

	b, err := automerge.As[bool](doc.Path("isValid").Get())
	if err != nil {
		panic(err)
	}
	fmt.Println("isValid:", b == true)

	v, err := automerge.As[string](doc.Path("foo", "bar").Get())
	if err != nil {
		panic(err)
	}
	fmt.Println("foo-bar:", v)

	type S struct {
		IsValid bool `automerge:"isValid"`
	}
	s, err := automerge.As[*S](doc.Root())
	if err != nil {
		panic(err)
	}
	fmt.Println("root valid:", s.IsValid == true)

	// Output:
	// isValid: true
	// foo-bar: baz
	// root valid: true
}

func ExampleSyncState() {
	doc := automerge.New()
	syncState := automerge.NewSyncState(doc)

	docUpdated := make(chan bool)
	recv := make(chan []byte)
	send := make(chan []byte)

loop:
	// generate an initial message, and then do so again
	// after receiving updates from the peer or making local changes
	for {
		msg, valid := syncState.GenerateMessage()
		if valid {
			send <- msg.Bytes()
		}

		select {
		case msg, ok := <-recv:
			if !ok {
				break loop
			}

			_, err := syncState.ReceiveMessage(msg)
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

/*
func ExampleDoc_SaveIncremental() {
	doc1 := automerge.New()
	// make initial changes

	changes := make(chan []byte)

	go func() {
		doc2 := automerge.Load(initialState)

		for ch := range changes {
			err := doc2.LoadIncremental(ch)
			if err != nil {
				panic(err)
			}
		}
	}()

	for {

	}
}
*/
