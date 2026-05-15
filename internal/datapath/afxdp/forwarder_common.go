package afxdp

import (
	"fmt"

	"smip-mwp/internal/routing"
	"smip-mwp/internal/wire"
)

// PrepareForPacket parses a header buffer and returns the selected next-hop ID
// and queue for forwarding using the routing table. This is a small, testable
// unit that real AF_XDP forwarder code will call for steering decisions.
func PrepareForPacket(buf []byte, rt *routing.Table) (nextHop [32]byte, queue int, err error) {
	h, err := wire.ParseHeader(buf)
	if err != nil {
		return [32]byte{}, 0, fmt.Errorf("parse header: %w", err)
	}

	nh, ok := rt.LookupOrPredict(h.SrcID, h.DstID, h.FlowLabel)
	if !ok {
		return [32]byte{}, 0, fmt.Errorf("no next hop available")
	}

	// For now, choose queue 0 for all routes. Later this will consult Router policy.
	return nh, 0, nil
}
