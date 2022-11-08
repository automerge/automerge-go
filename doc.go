/*
Package automerge provides the ability to interact with automerge documents.
It is a featureful wrapper around automerge-rs that uses Cgo to avoid reimplementing
the core engine from scratch.

Automerge documents have a JSON-like structure, they consist of a root map which
has string keys, and values may have any of the primative types:

* bool, string, []byte, float64, (untyped) nil, int64, uint64, time.Time

Additionally values may have any of the four special automerge types:

* Map - a mutable map[string]any
* List - a mutable []any
* Text – a mutable string
* Counter – a mutable int64

Trying to access a value that is not set will return a Value with KindVoid

Automerge-go will convert your go value to the most appropriate type if
possible, and error if not. For example structs are maps are converted to
*automerge.Map, slices and arrays to *List, int, int32 are converted to float64

Depending on your use-case there are a few ways to interact with the document,
the recommended approach for reading is to cast the document to a go value:

	doc, err := automerge.Load(bytes)
	if err != nil { return err }

	myVal, err := automerge.As[*myType](doc.Get())
	if err != nil { return err }

If you wish to access data from within the document, or modify data, the best
way is to use a path:

	err := doc.Path("x", "y", 0).Set(6)
	v, err := automerge.As[int](doc.Path("x", "y", 0).Get())

The automerge types have additional methods (beyond just "Get" and "Set"). You
can get access to an automerge type either by casting the value explicitly:

	map, err := automerge.As[*automerge.Map](doc.Path("x").Get())
	iter := map.Iter()

Or for convenience, by directly using the path. When you use this approach
automerge-go will create the object (if it doesn't exist already) the first time
you write to it. Read-only methods will not modify the document, but may fail
if the value referenced by the path doesn't have the correct type, or return void to
indicate that the path you're accessing does not exist in the document.

	iter := doc.Path("collection").Map().Iter()
	for {
		k, v, valid := iter.Next()
		if !valid { break }
	}
	if err := iter.Error(); err != nil {
		return err
	}
*/
package automerge
