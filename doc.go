package automerge

// #include "automerge.h"
import "C"
import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

type actorID struct {
	item     *item
	cActorID *C.AMactorId
}

func (ai *actorID) String() string {
	defer runtime.KeepAlive(ai)
	return fromByteSpanStr(C.AMactorIdStr(ai.cActorID))
}

// NewActorID generates a new unique actor id.
func NewActorID() string {
	return must(wrap(C.AMactorIdInit()).item()).actorID().String()
}

// Doc represents an automerge document. You can read and write the
// values of the document with [Doc.Root], [Doc.RootMap] or [Doc.Path],
// and other methods are provided to enable collaboration and accessing
// historical data.
// After writing to the document you should immediately call [Doc.Commit] to
// explicitly create a [Change], though if you forget to do this most methods
// on a document will create an anonymous change on your behalf.
type Doc struct {
	item *item
	cDoc *C.AMdoc

	m sync.Mutex
}

func (d *Doc) lock() (*C.AMdoc, func()) {
	d.m.Lock()
	locked := true
	return d.cDoc, func() {
		if locked {
			locked = false
			d.m.Unlock()
		}
	}
}

// New creates a new empty document
func New() *Doc {
	return must(wrap(C.AMcreate(nil)).item()).doc()
}

// Load loads a document from its serialized form
func Load(b []byte) (*Doc, error) {
	cbytes, free := toByteSpan(b)
	defer free()

	item, err := wrap(C.AMload(cbytes.src, cbytes.count)).item()
	if err != nil {
		return nil, err
	}
	return item.doc(), nil
}

// Save exports a document to its serialized form
func (d *Doc) Save() []byte {
	cDoc, unlock := d.lock()
	defer unlock()

	return must(wrap(C.AMsave(cDoc)).item()).bytes()
}

// RootMap returns the root of the document as a Map
func (d *Doc) RootMap() *Map {
	return &Map{doc: d, objID: rootObjID}
}

// Root returns the root of the document as a Value
// of [KindMap]
func (d *Doc) Root() *Value {
	return &Value{kind: KindMap, doc: d, val: d.RootMap()}
}

// Path returns a [*Path] that points to a position in the doc.
// Path will panic unless each path component is a string or an int.
// Calling Path with no arguments returns a path to the [Doc.Root].
func (d *Doc) Path(path ...any) *Path {
	return (&Path{d: d}).Path(path...)
}

// CommitOptions are (rarer) options passed to commit.
// If Time is not set then time.Now() is used. To omit a timestamp pass a pointer to the zero time: &time.Time{}
// If AllowEmpty is not set then commits with no operations will error.
type CommitOptions struct {
	Time       *time.Time
	AllowEmpty bool
}

// Commit adds a new version to the document with all operations so far.
// The returned ChangeHash is the new head of the document.
// Note: You should call commit immediately after modifying the document
// as most methods that inspect or modify the documents' history
// will automatically commit any outstanding changes.
func (d *Doc) Commit(msg string, opts ...CommitOptions) (ChangeHash, error) {
	cDoc, unlock := d.lock()
	defer unlock()

	allowEmpty := false
	time := time.Now()
	for _, o := range opts {
		if o.AllowEmpty {
			allowEmpty = true
		}
		if o.Time != nil {
			time = *o.Time
		}
	}

	millis := (*C.int64_t)(C.NULL)
	if !time.IsZero() {
		m := time.UnixMilli()
		millis = (*C.int64_t)(&m)
	}

	cMsg := C.AMbyteSpan{src: (*C.uchar)(C.NULL), count: 0}
	if msg != "" {
		var free func()
		cMsg, free = toByteSpanStr(msg)
		defer free()
	}

	item, err := wrap(C.AMcommit(cDoc, cMsg, millis)).item()
	if err != nil {
		return ChangeHash{}, err
	}
	if err == nil && item.Kind() == KindVoid {
		if !allowEmpty {
			return ChangeHash{}, fmt.Errorf("Commit is empty")
		}
		item, err = wrap(C.AMemptyChange(cDoc, cMsg, millis)).item()
		if err != nil {
			return ChangeHash{}, err
		}
	}

	return item.changeHash(), nil
}

// Heads returns the hashes of the current heads for the document.
// For a new document with no changes, this will have length zero.
// If you have just created a commit, this will have length one. If
// you have applied independent changes from multiple actors, then the
// length will be greater that one.
// If you'd like to merge independent changes together call [Doc.Commit]
// passing a [CommitOptions] with AllowEmpty set to true.
func (d *Doc) Heads() []ChangeHash {
	cDoc, unlock := d.lock()
	defer unlock()

	items := must(wrap(C.AMgetHeads(cDoc)).items())
	return mapItems(items, func(i *item) ChangeHash {
		return i.changeHash()
	})
}

// Change gets a specific change by hash.
func (d *Doc) Change(ch ChangeHash) (*Change, error) {
	cDoc, unlock := d.lock()
	defer unlock()

	byteSpan, free := toByteSpan(ch[:])
	defer free()

	item, err := wrap(C.AMgetChangeByHash(cDoc, byteSpan.src, byteSpan.count)).item()
	if err != nil {
		return nil, err
	}

	if item.Kind() == KindVoid {
		return nil, fmt.Errorf("hash %s does not correspond to a change in this document", ch)
	}

	return item.change(), nil
}

// Changes returns all changes made to the doc since the given heads.
// If since is empty, returns all changes to recreate the document.
func (d *Doc) Changes(since ...ChangeHash) ([]*Change, error) {
	cDoc, unlock := d.lock()
	defer unlock()

	items, err := itemsFromChangeHashes(since)
	if err != nil {
		return nil, err
	}
	cSince, free := createItems(items)
	defer free()

	items, err = wrap(C.AMgetChanges(cDoc, cSince)).items()
	if err != nil {
		return nil, err
	}
	return mapItems(items, func(i *item) *Change {
		return i.change()
	}), nil
}

// Apply the given change(s) to the document
func (d *Doc) Apply(chs ...*Change) error {
	if len(chs) == 0 {
		return nil
	}

	items := []*item{}
	for _, ch := range chs {
		items = append(items, ch.item)
	}

	cDoc, unlock := d.lock()
	defer unlock()
	cChs, free := createItems(items)
	defer free()

	return wrap(C.AMapplyChanges(cDoc, cChs)).void()
}

// SaveIncremental exports the changes since the last call to [Doc.Save] or
// [Doc.SaveIncremental] for passing to [Doc.LoadIncremental] on a different doc.
// See also [SyncState] for a more managed approach to syncing.
func (d *Doc) SaveIncremental() []byte {
	cDoc, unlock := d.lock()
	defer unlock()

	return must(wrap(C.AMsaveIncremental(cDoc)).item()).bytes()
}

// LoadIncremental applies the changes exported by [Doc.SaveIncremental].
// It is the callers responsibility to ensure that every incremental change
// is applied to keep the documents in sync.
// See also [SyncState] for a more managed approach to syncing.
func (d *Doc) LoadIncremental(raw []byte) error {
	cDoc, unlock := d.lock()
	defer unlock()
	cBytes, free := toByteSpan(raw)
	defer free()

	// returns the number of bytes read...
	_, err := wrap(C.AMloadIncremental(cDoc, cBytes.src, cBytes.count)).item()
	return err
}

// Fork returns a new, independent, copy of the document
// if asOf is empty then it is forked in its current state.
// otherwise it returns a version as of the given heads.
func (d *Doc) Fork(asOf ...ChangeHash) (*Doc, error) {
	items, err := itemsFromChangeHashes(asOf)
	if err != nil {
		return nil, err
	}

	cDoc, unlock := d.lock()
	defer unlock()
	cAsOf, free := createItems(items)
	defer free()

	item, err := wrap(C.AMfork(cDoc, cAsOf)).item()
	if err != nil {
		return nil, err
	}
	return item.doc(), nil
}

// Merge extracts all changes from d2 that are not in d
// and then applies them to d.
func (d *Doc) Merge(d2 *Doc) ([]ChangeHash, error) {
	cDoc, unlock := d.lock()
	defer unlock()
	cDoc2, unlock2 := d2.lock()
	defer unlock2()

	items, err := wrap(C.AMmerge(cDoc, cDoc2)).items()
	if err != nil {
		return nil, err
	}
	return mapItems(items, func(i *item) ChangeHash { return i.changeHash() }), nil
}

// ActorID returns the current actorId of the doc hex-encoded
// This is used for all operations that write to the document.
// By default a random ActorID is generated, but you can customize
// this with [Doc.SetActorID].
func (d *Doc) ActorID() string {
	cDoc, unlock := d.lock()
	defer unlock()

	return must(wrap(C.AMgetActorId(cDoc)).item()).actorID().String()
}

// SetActorID updates the current actorId of the doc.
// Valid actor IDs are a string with an even number of hex-digits.
func (d *Doc) SetActorID(id string) error {
	ai, err := itemFromActorID(id)
	if err != nil {
		return err
	}

	cDoc, unlock := d.lock()
	defer unlock()
	defer runtime.KeepAlive(ai)

	return wrap(C.AMsetActorId(cDoc, ai.actorID().cActorID)).void()
}
