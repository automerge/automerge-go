package automerge

import (
	"fmt"
	"runtime"
	"time"
	"unsafe"
)

// #include <automerge.h>
import "C"

// Kind represents the underlying type of a Value
type Kind uint

var (
	// KindVoid indicates the value was not present
	KindVoid Kind = C.AM_VAL_TYPE_VOID
	// KindBool indicates a bool
	KindBool Kind = C.AM_VAL_TYPE_BOOL
	// KindBytes indicates a []byte
	KindBytes Kind = C.AM_VAL_TYPE_BYTES
	// KindCounter indicates an *automerge.Counter
	KindCounter Kind = C.AM_VAL_TYPE_COUNTER
	// KindFloat64 indicates a float64
	KindFloat64 Kind = C.AM_VAL_TYPE_F64
	// KindInt indicates an int
	KindInt64 Kind = C.AM_VAL_TYPE_INT
	// KindUint indicates a uint
	KindUint64 Kind = C.AM_VAL_TYPE_UINT
	// KindNull indicates an explicit null was present
	KindNull Kind = C.AM_VAL_TYPE_NULL
	// KindStr indicates a string
	KindStr Kind = C.AM_VAL_TYPE_STR
	// KindTime indicates a time.Time
	KindTime Kind = C.AM_VAL_TYPE_TIMESTAMP
	// KindUnknown indicates an unknown type from a future version of automerge
	KindUnknown Kind = C.AM_VAL_TYPE_UNKNOWN

	// KindMap indicates an *automerge.Map
	KindMap Kind = kindObjType | C.AM_OBJ_TYPE_MAP
	// KindList indicates an *automerge.List
	KindList Kind = kindObjType | C.AM_OBJ_TYPE_LIST
	// KindText indicates an *automerge.Text
	KindText Kind = kindObjType | C.AM_OBJ_TYPE_TEXT

	kindActorID     Kind = C.AM_VAL_TYPE_ACTOR_ID
	kindDoc         Kind = C.AM_VAL_TYPE_DOC
	kindChange      Kind = C.AM_VAL_TYPE_CHANGE
	kindChangeHash  Kind = C.AM_VAL_TYPE_CHANGE_HASH
	kindObjType     Kind = C.AM_VAL_TYPE_OBJ_TYPE
	kindSyncState   Kind = C.AM_VAL_TYPE_SYNC_STATE
	kindSyncMessage Kind = C.AM_VAL_TYPE_SYNC_MESSAGE
	kindMark        Kind = C.AM_VAL_TYPE_MARK
)

var kindDescriptions = map[Kind]string{
	KindVoid:        "KindVoid",
	KindBool:        "KindBool",
	KindBytes:       "KindBytes",
	KindCounter:     "KindCounter",
	KindFloat64:     "KindFloat64",
	KindInt64:       "KindInt64",
	KindUint64:      "KindUint64",
	KindNull:        "KindNull",
	KindStr:         "KindStr",
	KindTime:        "KindTime",
	KindUnknown:     "KindUnknown",
	KindMap:         "KindMap",
	KindList:        "KindList",
	KindText:        "KindText",
	kindActorID:     "kindActorID",
	kindDoc:         "kindDoc",
	kindChange:      "kindChange",
	kindChangeHash:  "kindChangeHash",
	kindObjType:     "kindObjType",
	kindSyncState:   "kindSyncState",
	kindSyncMessage: "kindSyncMessage",
	kindMark:        "KindMark",
}

// String returns a human-readable representation of the Kind
func (k Kind) String() string {
	if s, ok := kindDescriptions[k]; ok {
		return s
	}
	return fmt.Sprintf("Kind(%v)", uint(k))
}

// item wraps an AMitem
type item struct {
	result *result
	cItem  *C.AMitem

	kind Kind
}

func (i *item) Kind() Kind {
	if i.kind == 0 {
		defer runtime.KeepAlive(i)
		i.kind = Kind(C.AMitemValType(i.cItem))
	}
	return i.kind
}

func (i *item) failCast(k Kind) {
	panic(fmt.Errorf("automerge: expected item with %v, got %v", k, i.Kind()))
}

func (i *item) actorID() *actorID {
	defer runtime.KeepAlive(i)

	ai := &actorID{item: i}
	if !C.AMitemToActorId(i.cItem, &ai.cActorID) {
		i.failCast(kindActorID)
	}
	return ai
}

func (i *item) bool() (ret bool) {
	defer runtime.KeepAlive(i)

	if !C.AMitemToBool(i.cItem, (*C.bool)(&ret)) {
		i.failCast(KindBool)
	}
	return ret
}

func (i *item) str() string {
	defer runtime.KeepAlive(i)

	var bs C.AMbyteSpan
	if !C.AMitemToStr(i.cItem, &bs) {
		i.failCast(KindBytes)
	}
	return fromByteSpanStr(bs)
}

func (i *item) bytes() []byte {
	defer runtime.KeepAlive(i)

	var bs C.AMbyteSpan
	if !C.AMitemToBytes(i.cItem, &bs) {
		i.failCast(KindBytes)
	}
	return fromByteSpan(bs)
}

func (i *item) change() *Change {
	defer runtime.KeepAlive(i)

	var c *C.AMchange
	if !C.AMitemToChange(i.cItem, &c) {
		i.failCast(kindChange)
	}
	return &Change{item: i, cChange: c}
}

func (i *item) changeHash() ChangeHash {
	defer runtime.KeepAlive(i)

	var bs C.AMbyteSpan
	if !C.AMitemToChangeHash(i.cItem, &bs) {
		i.failCast(kindChangeHash)
	}
	return *(*ChangeHash)(fromByteSpan(bs))
}

func (i *item) doc() *Doc {
	defer runtime.KeepAlive(i)

	d := &Doc{item: i}
	if !C.AMitemToDoc(i.cItem, &d.cDoc) {
		i.failCast(kindDoc)
	}
	return d
}

func (i *item) float64() (ret float64) {
	defer runtime.KeepAlive(i)

	if !C.AMitemToF64(i.cItem, (*C.double)(&ret)) {
		i.failCast(KindFloat64)
	}
	return ret
}

func (i *item) int64() (ret int64) {
	defer runtime.KeepAlive(i)

	if !C.AMitemToInt(i.cItem, (*C.int64_t)(&ret)) {
		i.failCast(KindInt64)
	}
	return ret
}

func (i *item) uint64() (ret uint64) {
	defer runtime.KeepAlive(i)

	if !C.AMitemToUint(i.cItem, (*C.uint64_t)(&ret)) {
		i.failCast(KindUint64)
	}
	return ret
}

func (i *item) time() time.Time {
	defer runtime.KeepAlive(i)

	var ms int64
	if !C.AMitemToTimestamp(i.cItem, (*C.int64_t)(&ms)) {
		i.failCast(KindTime)
	}
	return time.UnixMilli(ms)
}

func (i *item) counter() *Counter {
	defer runtime.KeepAlive(i)

	var val int64
	if !C.AMitemToCounter(i.cItem, (*C.int64_t)(&val)) {
		i.failCast(KindCounter)
	}
	return &Counter{val: val}
}

func (i *item) mapKey() string {
	defer runtime.KeepAlive(i)

	var bs C.AMbyteSpan
	if !C.AMitemKey(i.cItem, &bs) {
		panic(fmt.Errorf("expected C.AM_IDX_TYPE_KEY, got %v", C.AMitemIdxType(i.cItem)))
	}
	return fromByteSpanStr(bs)
}

func (i *item) objID() *objID {
	defer runtime.KeepAlive(i)

	oi := C.AMitemObjId(i.cItem)
	if oi == nil {
		return nil
	}
	return &objID{item: i, cObjID: oi}
}

func (i *item) syncState() *SyncState {
	defer runtime.KeepAlive(i)

	ss := &SyncState{item: i}
	if !C.AMitemToSyncState(i.cItem, &ss.cSyncState) {
		i.failCast(kindSyncState)
	}
	return ss
}

func (i *item) syncMessage() *syncMessage {
	defer runtime.KeepAlive(i)

	ss := &syncMessage{item: i}
	if !C.AMitemToSyncMessage(i.cItem, &ss.cSyncMessage) {
		if i.Kind() == KindVoid {
			return nil
		}
		i.failCast(kindSyncMessage)
	}
	return ss
}

type objID struct {
	item *item

	cObjID *C.AMobjId
}

func (o *objID) objKind(d *Doc) Kind {
	// not using d.Lock() here because the doc
	// is likely already locked by List.Get()/Map.Get()
	defer runtime.KeepAlive(d)
	defer runtime.KeepAlive(o)
	return Kind(C.AMobjObjType(d.cDoc, o.cObjID)) | kindObjType
}

func itemFromActorID(id string) (*item, error) {
	bytes, free := toByteSpanStr(id)
	defer free()

	return wrap(C.AMactorIdFromStr(bytes)).item()
}

func itemsFromChangeHashes(ch []ChangeHash) ([]*item, error) {
	items := []*item{}
	for _, c := range ch {
		bytes, free := toByteSpan(c[:])
		defer free()
		item, err := wrap(C.AMitemFromChangeHash(bytes)).item()
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func createItems(is []*item) (*C.AMitems, func()) {
	if len(is) == 0 {
		return nil, func() {}
	}

	result := is[0].result
	for _, i := range is[1:] {
		result = wrap(C.AMresultCat(result.cResult, i.result.cResult))
	}
	ret := C.AMresultItems(result.cResult)
	return &ret, func() {
		runtime.KeepAlive(result)
	}
}

func mapItems[T any](is []*item, f func(i *item) T) []T {
	ret := []T{}
	for _, i := range is {
		ret = append(ret, f(i))
	}
	return ret
}

func toByteSpan(b []byte) (C.AMbyteSpan, func()) {
	if len(b) == 0 {
		return C.AMbyteSpan{src: (*C.uchar)(C.NULL), count: 0}, func() {}
	}
	return C.AMbyteSpan{src: (*C.uchar)(&b[0]), count: C.size_t(len(b))}, func() {
		runtime.KeepAlive(b)
	}
}

func fromByteSpan(bs C.AMbyteSpan) []byte {
	return C.GoBytes(unsafe.Pointer(bs.src), C.int(bs.count))
}

func toByteSpanStr(s string) (C.AMbyteSpan, func()) {
	return toByteSpan([]byte(s))
}

func fromByteSpanStr(bs C.AMbyteSpan) string {
	return string(fromByteSpan(bs))
}
