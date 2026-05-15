package afxdp

import (
	"testing"

	"smip-mwp/internal/routing"
	"smip-mwp/internal/wire"
)

func TestPrepareForPacket(t *testing.T) {
	var src, dst [32]byte
	copy(src[:], []byte("src-prepare-test-000000000000000"))
	copy(dst[:], []byte("dst-prepare-test-000000000000000"))

	h := wire.Header{
		SrcID:     src,
		DstID:     dst,
		FlowLabel: 0x1010,
		SeqNum:    1,
		Length:    0,
	}
	buf := wire.NewHeaderBuffer(int(h.Length))
	if err := h.Marshal(buf); err != nil {
		t.Fatalf("marshal: %v", err)
	}

	rt := routing.NewTable()
	var next [32]byte
	copy(next[:], []byte("next-prepare-test-00000000000000"))
	rt.UpdateRoute(routing.RouteEntry{DestID: dst, NextHopID: next})

	nh, q, err := PrepareForPacket(buf, rt)
	if err != nil {
		t.Fatalf("prepare: %v", err)
	}
	if nh != next {
		t.Fatalf("next hop mismatch")
	}
	if q != 0 {
		t.Fatalf("queue mismatch: got %d", q)
	}
}
