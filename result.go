package automerge

/*
#include "automerge.h"

#cgo LDFLAGS: -L${SRCDIR}/deps
#cgo darwin,arm64 LDFLAGS: -lautomerge_core_darwin_arm64
#cgo darwin,amd64 LDFLAGS: -lautomerge_core_darwin_amd64
#cgo linux,arm64 LDFLAGS: -lautomerge_core_linux_arm64 -lm
#cgo linux,amd64 LDFLAGS: -lautomerge_core_linux_amd64 -lm
*/
import "C"
import (
	"fmt"
	"runtime"

	_ "github.com/automerge/automerge-go/deps"
)

// result wraps an AMresult, and arranges for it to be AMfree'd after
// the result is garbage collected.
type result struct {
	cResult *C.AMresult
}

func wrap(r *C.AMresult) *result {
	ret := &result{r}
	runtime.SetFinalizer(ret, func(*result) { C.AMresultFree(r) })
	return ret
}

func (r *result) void() error {
	item, err := r.item()
	if err != nil {
		return err
	}
	if err == nil && item.Kind() != KindVoid {
		return fmt.Errorf("expected KindVoid, got: %s", item.Kind())
	}
	return nil
}

func (r *result) item() (*item, error) {
	items, err := r.items()
	if err != nil {
		return nil, err
	}
	if len(items) != 1 {
		return nil, fmt.Errorf("automerge: expected single return value, got %v", len(items))
	}
	return items[0], nil
}

func (r *result) items() ([]*item, error) {
	defer runtime.KeepAlive(r)

	switch C.AMresultStatus(r.cResult) {
	case C.AM_STATUS_OK:
		items := C.AMresultItems(r.cResult)
		ret := []*item{}
		for {
			i := C.AMitemsNext(&items, 1)
			if i == nil {
				break
			}

			ret = append(ret, &item{result: r, cItem: i})
		}
		return ret, nil
	case C.AM_STATUS_ERROR:
		msg := fromByteSpanStr(C.AMresultError(r.cResult))
		return nil, fmt.Errorf(msg)
	case C.AM_STATUS_INVALID_RESULT:
		return nil, fmt.Errorf("automerge: invalid result")
	default:
		return nil, fmt.Errorf("automerge: invalid result status")
	}
}

func must[T any](r T, e error) T {
	if e != nil {
		panic(e)
	}
	return r
}
