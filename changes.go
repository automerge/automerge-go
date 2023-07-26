ackage automerge

// #include "automerge.h"
import "C"
import (
	"encoding/hex"
	"fmt"
	"runtime"
	"time"
)

// ChangeHash is a SHA-256 hash identifying an automerge change.
// Like a git commit, the hash encompasses both the changes made,
// any metadata (like commit message, or timestamp)
// and any changes on which this change depends.
type ChangeHash [32]byte

// String returns the hex-encoded form of the change hash
func (ch ChangeHash) String() string {
	return hex.EncodeToString(ch[:])
}

// NewChangeHash creates a change has from its hex representation.
func NewChangeHash(s string) (ChangeHash, error) {
	b, err := hex.DecodeString(s)
	if err != nil {
		return ChangeHash{}, err
	}
	if len(b) != 32 {
		return ChangeHash{}, fmt.Errorf("automerge.NewChangeHash: expected 32 bytes, got %v", len(b))
	}
	return *(*ChangeHash)(b), nil
}

// Change is a set of mutations to the document. It is analagous
// to a commit in a version control system like Git.
type Change struct {
	item    *item
	cChange *C.AMchange
}

// ActorID identifies the actor that made the change hex-encoded
func (c *Change) ActorID() string {
	defer runtime.KeepAlive(c)
	return must(wrap(C.AMchangeActorId(c.cChange)).item()).actorID().String()
}

// ActorSeq is 1 for the first change by a given
// actor, 2 for the next, and so on.
func (c *Change) ActorSeq() uint64 {
	defer runtime.KeepAlive(c)
	return uint64(C.AMchangeSeq(c.cChange))
}

// Hash identifies the change by the SHA-256 of its binary representation
func (c *Change) Hash() ChangeHash {
	defer runtime.KeepAlive(c)
	return *(*ChangeHash)(fromByteSpan(C.AMchangeHash(c.cChange)))
}

// Dependencies returns the hashes of all changes that this change
// directly depends on.
func (c *Change) Dependencies() []ChangeHash {
	defer runtime.KeepAlive(c)
	items := must(wrap(C.AMchangeDeps(c.cChange)).items())
	return mapItems(items, func(i *item) ChangeHash { return i.changeHash() })
}

// Message returns the commit message (if any)
func (c *Change) Message() string {
	defer runtime.KeepAlive(c)
	return fromByteSpanStr(C.AMchangeMessage(c.cChange))
}

// Timestamp returns the commit time (or the zero time if one was not set)
func (c *Change) Timestamp() time.Time {
	defer runtime.KeepAlive(c)
	return time.UnixMilli(int64(C.AMchangeTime(c.cChange)))
}

// Save exports the change for transferring between systems
func (c *Change) Save() []byte {
	defer runtime.KeepAlive(c)
	return fromByteSpan(C.AMchangeRawBytes(c.cChange))
}

// LoadChanges loads changes from bytes (see also [SaveChanges] and [Change.Save])
func LoadChanges(raw []byte) ([]*Change, error) {
	cBytes, free := toByteSpan(raw)
	defer free()

	ret, err := wrap(C.AMchangeLoadDocument(cBytes.src, cBytes.count)).items()
	if err != nil {
		return nil, err
	}
	return mapItems(ret, func(i *item) *Change { return i.change() }), nil
}

// SaveChanges saves multiple changes to bytes (see also [LoadChanges])
func SaveChanges(cs []*Change) []byte {
	out := []byte{}

	for _, c := range cs {
		out = append(out, c.Save()...)
	}
	return out
}
