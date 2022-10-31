package automerge

//go:generate go run generate/gen.go

/*
#cgo CFLAGS: -I${SRCDIR}/deps/include -Wall
#cgo LDFLAGS: -lautomerge
#cgo darwin,arm64 LDFLAGS: -L${SRCDIR}/deps/darwin_arm64

#include "automerge.h"
#include <stdlib.h>

AMvalueVariant AMresultValueTag(AMresult *r) { return AMresultValue(r).tag; }
const struct AMactorId *AMresultValueActorId(AMresult *r) { return AMresultValue(r).actor_id; }
bool AMresultValueBoolean(AMresult *r) { return AMresultValue(r).boolean; }
struct AMbyteSpan AMresultValueBytes(AMresult *r) { return AMresultValue(r).bytes; }
struct AMchangeHashes AMresultValueChangeHashes(AMresult *r) { return AMresultValue(r).change_hashes; }
struct AMchanges AMresultValueChanges(AMresult *r) { return AMresultValue(r).changes; }
int64_t AMresultValueCounter(AMresult *r) { return AMresultValue(r).counter; }
struct AMdoc *AMresultValueDoc(AMresult *r) { return AMresultValue(r).doc; }
double AMresultValueF64(AMresult *r) { return AMresultValue(r).f64; }
int64_t AMresultValueInt(AMresult *r) { return AMresultValue(r).int_; }
struct AMlistItems AMresultValueListItems(AMresult *r) { return AMresultValue(r).list_items; }
struct AMmapItems AMresultValueMapItems(AMresult *r) { return AMresultValue(r).map_items; }
const struct AMobjId *AMresultValueObjId(AMresult *r) { return AMresultValue(r).obj_id; }
struct AMobjItems AMresultValueObjItems(AMresult *r) { return AMresultValue(r).obj_items; }
const char *AMresultValueStr(AMresult *r) { return AMresultValue(r).str; }
struct AMstrs AMresultValueStrs(AMresult *r) { return AMresultValue(r).strs; }
const struct AMsyncMessage *AMresultValueSyncMessage(AMresult *r) { return AMresultValue(r).sync_message; }
struct AMsyncState *AMresultValueSyncState(AMresult *r) { return AMresultValue(r).sync_state; }
int64_t AMresultValueTimestamp(AMresult *r) { return AMresultValue(r).timestamp; }
uint64_t AMresultValueUint(AMresult *r) { return AMresultValue(r).uint; }
struct AMunknownValue AMresultValueUnknown(AMresult *r) { return AMresultValue(r).unknown; }
*/
import "C"
import (
	"fmt"
	"runtime"
	"unsafe"
)

// ActorId represents an actor working on the document.
type ActorId struct {
	r *C.AMresult
	v *C.AMactorId
}

func (a *ActorId) init(r *C.AMresult) error {
	if tag := C.AMresultValueTag(r); tag != C.AM_VALUE_ACTOR_ID {
		return fmt.Errorf("expected VALUE_ACTOR_ID, got %v", tag)
	}

	a.r = r
	a.v = C.AMresultValueActorId(r)
	return nil
}

func NewActorId() (*ActorId, error) {
	return call[*ActorId](C.AMactorIdInit())
}

func ActorIdFromString(id string) (*ActorId, error) {
	cstr := C.CString(id)
	defer C.free(unsafe.Pointer(cstr))
	return call[*ActorId](C.AMactorIdInitStr(cstr))
}

func ActorIdFromBytes(id []byte) (*ActorId, error) {
	cbytes := C.CBytes(id)
	defer C.free(cbytes)
	return call[*ActorId](C.AMactorIdInitBytes((*C.uchar)(cbytes), C.ulong(len(id))))
}

func (a *ActorId) Bytes() []byte {
	defer runtime.KeepAlive(a)
	ret := C.AMactorIdBytes(a.v)
	return C.GoBytes(unsafe.Pointer(ret.src), C.int(ret.count))
}

func (a *ActorId) String() string {
	defer runtime.KeepAlive(a)
	return C.GoString(C.AMactorIdStr(a.v))
}

func (a *ActorId) Cmp(b *ActorId) int {
	defer runtime.KeepAlive(a)
	defer runtime.KeepAlive(b)
	return int(C.AMactorIdCmp(a.v, b.v))
}

type Doc struct {
	*Map

	r *C.AMresult
	v *C.AMdoc
}

func (d *Doc) init(r *C.AMresult) error {
	if tag := C.AMresultValueTag(r); tag != C.AM_VALUE_DOC {
		return fmt.Errorf("expected VALUE_DOC, got %v", tag)
	}

	d.r = r
	d.v = C.AMresultValueDoc(r)
	d.Map = &Map{d: d, o: &objId{v: (*C.AMobjId)(C.AM_ROOT)}}
	return nil
}

func New(actorId *ActorId) (*Doc, error) {
	var a *C.AMactorId
	if actorId != nil {
		defer runtime.KeepAlive(actorId)
		a = actorId.v
	}

	return call[*Doc](C.AMcreate(a))
}

type objId struct {
	r *C.AMresult
	v *C.AMobjId
}

func (o *objId) init(r *C.AMresult) error {
	if tag := C.AMresultValueTag(r); tag != C.AM_VALUE_OBJ_ID {
		return fmt.Errorf("expected VALUE_OBJ_ID, got %v", tag)
	}

	o.r = r
	o.v = C.AMresultValueObjId(r)
	return nil
}

type Map struct {
	d *Doc
	o *objId
}

type void struct{}

func (m *Map) CreateMap(key string) (*Map, error) {
	defer runtime.KeepAlive(m)
	cstr := C.CString(key)
	defer C.free(unsafe.Pointer(cstr))

	objId, err := call[*objId](C.AMmapPutObject(m.d.v, m.o.v, cstr, C.AM_OBJ_TYPE_MAP))
	if err != nil {
		return nil, err
	}

	return &Map{d: m.d, o: objId}, nil
}

type MapItems struct {
	r *C.AMresult
	v C.AMmapItems

	err error
}

func (mi *MapItems) init(r *C.AMresult) error {
	if tag := C.AMresultValueTag(r); tag != C.AM_VALUE_MAP_ITEMS {
		return fmt.Errorf("expected VALUE_MAP_ITEMS, got %v", tag)
	}

	mi.r = r
	mi.v = C.AMresultValueMapItems(r)
	return nil
}

func (mi *MapItems) Next() (string, *Value, bool) {
	if mi.err != nil {
		return "", nil, false
	}
	defer runtime.KeepAlive(mi)

	item := C.AMmapItemsNext(&mi.v, 1)
	if item == nil {
		return "", nil, false
	}

	key := C.GoString(C.AMmapItemKey(item))

	// TODO: extract value...
	var value *Value

	return key, value, true
}

func (m *Map) Iter() *MapItems {
	defer runtime.KeepAlive(m)

	iter, err := call[*MapItems](C.AMmapRange(m.d.v, m.o.v, nil, nil, nil))
	if err != nil {
		return &MapItems{err: err}
	}
	return iter
}

type Value struct {
}

type List struct {
	d *Doc
	o *objId
}

type initer interface {
	init(r *C.AMresult) error
}

func call[T interface {
	*X
	initer
}, X any](r *C.AMresult) (ret T, err error) {
	switch C.AMresultStatus(r) {
	case C.AM_STATUS_OK:
		ret = T(new(X))
		err = ret.init(r)
	case C.AM_STATUS_ERROR:
		err = fmt.Errorf(C.GoString(C.AMerrorMessage(r)))
	case C.AM_STATUS_INVALID_RESULT:
		err = fmt.Errorf("automerge: invalid result")
	default:
		err = fmt.Errorf("automerge: invalid result status")
	}

	if err != nil {
		C.AMfree(r)
		return nil, err
	}

	runtime.SetFinalizer(ret, func(ret T) { C.AMfree(r) })
	return ret, nil
}
