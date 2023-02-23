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
	doc  *Doc
	item *item

	cSyncState *C.AMsyncState
}

// NewSyncState returns a new sync state to sync with a peer
func NewSyncState(d *Doc) *SyncState {
	ss := must(wrap(C.AMsyncStateInit()).item()).syncState()
	ss.doc = d
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
	ss.doc = d
	return ss, nil
}

// ReceiveMessage should be called with every message created by GenerateMessage
// on the peer side.
func (ss *SyncState) ReceiveMessage(msg []byte) error {
	sm, err := loadSyncMessage(msg)
	if err != nil {
		return err
	}

	defer runtime.KeepAlive(ss)
	defer runtime.KeepAlive(sm)
	cDoc, unlock := ss.doc.lock()
	defer unlock()

	return wrap(C.AMreceiveSyncMessage(cDoc, ss.cSyncState, sm.cSyncMessage)).void()
}

// GenerateMessage generates the next message to send to the client.
// If `valid` is false the clients are currently in sync and there are
// no more messages to send (until you either modify the underlying document)
func (ss *SyncState) GenerateMessage() (bytes []byte, valid bool) {
	defer runtime.KeepAlive(ss)
	cDoc, unlock := ss.doc.lock()
	defer unlock()

	sm := must(wrap(C.AMgenerateSyncMessage(cDoc, ss.cSyncState)).item()).syncMessage()

	if sm == nil {
		return nil, false
	}

	return sm.save(), true
}

// Save serializes the sync state so that you can resume it later.
// This is an optimization to reduce the number of round-trips required
// to get two peers in sync at a later date.
func (ss *SyncState) Save() ([]byte, error) {
	defer runtime.KeepAlive(ss)

	item, err := wrap(C.AMsyncStateEncode(ss.cSyncState)).item()
	if err != nil {
		return nil, err
	}
	return item.bytes(), nil
}

type syncMessage struct {
	item         *item
	cSyncMessage *C.AMsyncMessage
}

func loadSyncMessage(msg []byte) (*syncMessage, error) {
	cBytes, free := toByteSpan(msg)
	defer free()

	item, err := wrap(C.AMsyncMessageDecode(cBytes.src, cBytes.count)).item()
	if err != nil {
		return nil, err
	}
	return item.syncMessage(), nil
}

func (sm *syncMessage) save() []byte {
	if sm == nil {
		return nil
	}
	defer runtime.KeepAlive(sm)
	return must(wrap(C.AMsyncMessageEncode(sm.cSyncMessage)).item()).bytes()
}
