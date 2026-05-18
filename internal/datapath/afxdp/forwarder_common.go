// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright 2026 rwilliamspbg-ops
//
// This file is part of SMIP-MWP.
// SMIP-MWP is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation; either version 3 of the License, or (at your option) any later version.
// See the LICENSE file in the project root for details.

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
