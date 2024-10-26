package automerge

// #include "automerge.h"
import "C"

import (
	"runtime"
)

// SyncState represents the state of syncing between a local copy of
// a doc and a peer; and lets you optimize bandwidth used to ensure
// two docs are always in sync.
type SyncState struct {
	Doc  *Doc
	item *item

	cSyncState *C.AMsyncState
}

// NewSyncState returns a new sync state to sync with a peer
func NewSyncState(d *Doc) *SyncState {
	ss := must(wrap(C.AMsyncStateInit()).item()).syncState()
	ss.Doc = d
	return ss
}

// LoadSyncState lets you resume syncing with a peer from where you left off.
func LoadSyncState(d *Doc, raw []byte) (*SyncState, error) {
	cBytes, free := toByteSpan(raw)
	defer free()

	item, err := wrap(C.AMsyncStateDecode(cBytes.src, cBytes.count)).item()
	if err != nil {
		return nil, err
	}
	ss := item.syncState()
	ss.Doc = d
	return ss, nil
}

// ReceiveMessage should be called with every message created by GenerateMessage
// on the peer side.
func (ss *SyncState) ReceiveMessage(msg []byte) (*SyncMessage, error) {
	sm, err := LoadSyncMessage(msg)
	if err != nil {
		return nil, err
	}

	defer runtime.KeepAlive(ss)
	defer runtime.KeepAlive(sm)
	cDoc, unlock := ss.Doc.lock()
	defer unlock()

	return sm, wrap(C.AMreceiveSyncMessage(cDoc, ss.cSyncState, sm.cSyncMessage)).void()
}

// GenerateMessage generates the next message to send to the client.
// If `valid` is false the clients are currently in sync and there are
// no more messages to send (until you either modify the underlying document)
func (ss *SyncState) GenerateMessage() (sm *SyncMessage, valid bool) {
	defer runtime.KeepAlive(ss)
	cDoc, unlock := ss.Doc.lock()
	defer unlock()

	sm = must(wrap(C.AMgenerateSyncMessage(cDoc, ss.cSyncState)).item()).syncMessage()

	if sm == nil {
		return nil, false
	}

	return sm, true
}

// Save serializes the sync state so that you can resume it later.
// This is an optimization to reduce the number of round-trips required
// to get two peers in sync at a later date.
func (ss *SyncState) Save() []byte {
	defer runtime.KeepAlive(ss)

	return must(wrap(C.AMsyncStateEncode(ss.cSyncState)).item()).bytes()
}

// SyncMessage is sent between peers to keep copies of a document in sync.
type SyncMessage struct {
	item         *item
	cSyncMessage *C.AMsyncMessage
}

// LoadSyncMessage decodes a sync message from a byte slice for inspection.
func LoadSyncMessage(msg []byte) (*SyncMessage, error) {
	cBytes, free := toByteSpan(msg)
	defer free()

	item, err := wrap(C.AMsyncMessageDecode(cBytes.src, cBytes.count)).item()
	if err != nil {
		return nil, err
	}
	return item.syncMessage(), nil
}

// Changes returns any changes included in this SyncMessage
func (sm *SyncMessage) Changes() []*Change {
	defer runtime.KeepAlive(sm)

	items := must(wrap(C.AMsyncMessageChanges(sm.cSyncMessage)).items())

	return mapItems(items, func(i *item) *Change { return i.change() })
}

// Heads gives the heads of the peer that generated the SyncMessage
func (sm *SyncMessage) Heads() []ChangeHash {
	defer runtime.KeepAlive(sm)

	items := must(wrap(C.AMsyncMessageHeads(sm.cSyncMessage)).items())

	return mapItems(items, func(i *item) ChangeHash { return i.changeHash() })
}

// Bytes returns a representation for sending over the network.
func (sm *SyncMessage) Bytes() []byte {
	if sm == nil {
		return nil
	}
	defer runtime.KeepAlive(sm)
	return must(wrap(C.AMsyncMessageEncode(sm.cSyncMessage)).item()).bytes()
}
