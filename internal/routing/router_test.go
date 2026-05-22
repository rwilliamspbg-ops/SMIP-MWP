package routing

import (
    "testing"
)

func makeID(s string) [32]byte {
    var b [32]byte
    copy(b[:], []byte(s))
    return b
}

func TestLookupExactAndPredictive(t *testing.T) {
    tbl := NewTable()
    idA := makeID("dest-A-000000000000000000000")
    idB := makeID("dest-B-000000000000000000000")
    nextA := makeID("nexthop-A-0000000000000000000")
    nextB := makeID("nexthop-B-0000000000000000000")

    tbl.UpdateRoute(RouteEntry{DestID: idA, NextHopID: nextA})
    tbl.UpdateRoute(RouteEntry{DestID: idB, NextHopID: nextB})

    if nh, ok := tbl.LookupNextHop(idA, 1); !ok || nh != nextA {
        t.Fatalf("LookupNextHop exact failed: got=%x ok=%v", nh, ok)
    }

    // Predictive for an unknown dst should return one of the known next-hops
    unknown := makeID("dest-UNKNOWN-000000000000000000")
    if nh, ok := tbl.PredictiveNextHop(makeID("src-1"), unknown, 7); !ok {
        t.Fatalf("PredictiveNextHop returned not-ok")
    } else {
        if nh != nextA && nh != nextB {
            t.Fatalf("PredictiveNextHop returned unexpected nextHop: %x", nh)
        }
    }

    // LookupOrPredict should prefer exact match
    if nh, ok := tbl.LookupOrPredict(makeID("src-1"), idB, 9); !ok || nh != nextB {
        t.Fatalf("LookupOrPredict exact failed: got=%x ok=%v", nh, ok)
    }
}
