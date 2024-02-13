/*
Package automerge provides the ability to interact with [automerge] documents.
It is a featureful wrapper around [automerge-rs] that uses cgo to avoid reimplementing
the core engine from scratch.

# Document Structure and Types

Automerge documents have a JSON-like structure, they consist of a root map which
has string keys, and values of any supported types.

Supported types include several immutable primitive types:

  - bool
  - string, []byte
  - float64, int64, uint64
  - time.Time (in millisecond precision)
  - nil (untyped)

And four mutable automerge types:

  - [Map] - a mutable map[string]any
  - [List] - a mutable []any
  - [Text] – a mutable string
  - [Counter] – an int64 that is incremented (instead of overwritten) by collaborators

If you read part of the doc that has no value set, automerge-go will return a
Value with Kind() == KindVoid. You cannot create such a Value directly or write
one to the document.

On write automerge-go will attempt to convert provided values to the most
appropriate type, and error if that is not possible.  For example structs are
maps are converted to [*Map], slices and arrays to [*List], most numeric types
are converted to float64 (the default number type for automerge), with the
exception of int64 and uint64.

On read automerge-go will return a [*Value], and you can use [As] to convert this
to a more useful type.

# Interacting with the Document

Depending on your use-case there are a few ways to interact with the document,
the recommended approach for reading is to cast the document to a go value:

	doc, err := automerge.Load(bytes)
	if err != nil { return err }

	myVal, err := automerge.As[*myType](doc.Root())
	if err != nil { return err }

If you wish to modify the document, or access just a subset, use a Path:

	err := doc.Path("x", "y", 0).Set(&myStruct{Header: "h"})
	v, err := automerge.As[*myStruct](doc.Path("x", "y", 0).Get())

It is always recommended to write the smallest change to the document, as this
will improve the experience of other collaborative editors.

Writing to a path will create any intermediate Map or List objects needed,
Reading from a path will not, but may return a void Value if the intermediate
objects don't exist.

The automerge mutable types have additional methods. You can access these
methods by calling [Path.Map], [Path.List], [Path.Text] or [Path.Counter] which
assume the path is of the type you say it is:

	values, err := doc.Path("collection").Map().Values()
	fmt.Println(values)

When you do this, any errors caused by traversing the path will be returned from
methods called on the returned objects.

# Controling formatting of structs

By default automerge will convert your struct to a map. For each public field in the
struct (the name starts with an uppercase letter) automerge will add an entry to the
map with the name of the field and the fields value converted to an automerge value
recursively.

You can control this behaviour using struct tags:

	struct Example {
		Renamed bool `automerge:"newname"`
		Private bool `automerge:"-"`
	}

If the tag is present and equal to "-" the field will be ignored by automerge,
otherwise the fields name will be set to the value of the tag.

# Syncing and concurrency

You can access methods on [*Doc] from multiple goroutines and access is mediated
appropriately. For other types, you must provide your own syncronization, or
only use them from one goroutine at a time.

If you retain a Map, List, Counter, or Text object while the document is being
modified concurrently be aware that its value may change, or it may be deleted
from the document. A safer pattern is to fork the document, make the changes you
want, and then merge your changes back into the shared document.

There are a few different ways to keep distributed automerge docs in sync. If you're mostly
making changes in one place and syncing them to another, you can use [Doc.SaveIncremental] and
[Doc.LoadIncremental]

	//* process 1 *
	initBytes, err := doc.Save()
	for {
		// ... make changes ...
		bytes, err := doc.SaveIncremental()
		ch <- bytes
	}

	//* process 2*
	err := automerge.Load(initBytes)
	for bytes := range ch {
		err := doc.LoadIncremental(bytes)
	}

If both peers are making changes you can use a [SyncState] object to keep them
in sync.  This wraps an underlying efficient sync protocol to minimize both
round-trips and bandwidth used.

	//* process 1 *
	syncState := automerge.NewSyncState(doc)
	for {
		m, valid := syncState.GenerateMessage()
		if valid {
			sendCh <- m.Bytes()
		}

		msg := <-recvCh
		_, err := syncState.ReceiveMessage(msg)
		if err != nil {
			panic(err)
		}
	}

	//* process 2 *
	syncState := automerge.NewSyncState(doc)
	for {
		msg := <-sendCh
		_, err := syncState.ReceiveMessage(msg)
		if err != nil {
			panic(err)
		}

		m, valid := syncState.GenerateMessage()
		if valid {
			recvCh <- m.Bytes()
		}
	}

If you need more flexibility, you can use [Doc.Changes] and [Doc.Apply] to
manually track the changes you want to transfer. This puts more burden on
the implementor to ensure an absense of bugs.

	//* process 1 *
	heads := doc.Heads()
	for {
		// ... make changes ...
		changes, err := doc.Changes(heads)
		heads = doc.Heads()
		bytes := changes.Save()
		ch <- bytes
	}

	//* process 2 *
	for bytes := range ch {
		changes, err := automerge.LoadChanges(bytes)
		err := doc.Apply(changes)
	}

[automerge]: https://automerge.org
[automerge-rs]: https://github.com/automerge/automerge-rs
*/
package automerge
