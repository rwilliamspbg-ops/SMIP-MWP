package afxdp

import (
	"testing"
)

func TestShardIndexAndSessionLifecycle(t *testing.T) {
	f := &Forwarder{}
	f.initSessionShards()

	var sid [16]byte
	copy(sid[:], []byte("session-00000001"))
	idx := f.getShardIndex(sid)
	if idx < 0 || idx >= numSessionShards {
		t.Fatalf("shard index out of range: %d", idx)
	}

	sess := &Session{FlowLabel: 7}
	f.AddSession(sid, sess)
	if got := f.GetSession(sid); got == nil {
		t.Fatalf("expected session to be present")
	}

	f.RemoveSession(sid)
	if got := f.GetSession(sid); got != nil {
		t.Fatalf("expected session to be removed")
	}
}
