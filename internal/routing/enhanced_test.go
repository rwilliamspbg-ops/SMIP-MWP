package routing

import (
	"testing"
)

func TestEnhancedTableAdditiveBehavior(t *testing.T) {
	base := NewTable()
	et := NewEnhancedTable(base)

	// Add a base exact route
	dst := makeID("dst-base")
	nh := makeID("nh-base")
	base.UpdateRoute(RouteEntry{DestID: dst, NextHopID: nh})

	// LPM route that should not interfere with exact lookup for different dst
	prefix := []byte{0x01, 0x02, 0x03}
	var lpmNH [32]byte
	copy(lpmNH[:], []byte("nh-lpm"))
	et.AddLPMRoute(prefix, 24, lpmNH, 10)

	// Exact lookup via base should still return the base next-hop
	if got, ok := base.LookupNextHop(dst, 0); !ok || got != nh {
		t.Fatalf("base lookup changed after adding LPM: got=%x ok=%v", got, ok)
	}

	// LPM lookup for a destination matching the prefix should return lpmNH
	var dst2 [32]byte
	copy(dst2[:], append(prefix, make([]byte, 29)...))
	got, l, ok := et.LookupLPM(dst2)
	if !ok {
		t.Fatalf("expected LPM match")
	}
	if got != lpmNH || l != 24 {
		t.Fatalf("unexpected LPM result: got=%x len=%d", got, l)
	}
}
