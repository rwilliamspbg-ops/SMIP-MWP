// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright 2026 rwilliamspbg-ops
//
// This file is part of SMIP-MWP.
// SMIP-MWP is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation; either version 3 of the License, or (at your option) any later version.
// See the LICENSE file in the project root for details.

package afxdp

// XDPDescriptor represents a single packet descriptor in UMEM.
// Addr is the frame index (not byte offset) used by the forwarder to index
// into UMEM frame arrays. Len is the frame length in bytes.
type XDPDescriptor struct {
	Addr uint64 // frame index
	Len  uint32 // Packet length
	Pad  uint32 // Padding for alignment
}
