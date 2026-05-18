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
	"sync"
	"testing"
)

// BenchmarkPacketAllocate measures allocating packets with make() each time.
func BenchmarkPacketAllocate(b *testing.B) {
	size := 2048
	for i := 0; i < b.N; i++ {
		pkt := make([]byte, size)
		_ = pkt
	}
}

// BenchmarkPacketPool measures using a sync.Pool to reuse packet buffers.
func BenchmarkPacketPool(b *testing.B) {
	size := 2048
	pool := &sync.Pool{New: func() interface{} { buf := make([]byte, size); return &buf }}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pktPtr := pool.Get().(*[]byte)
		pkt := *pktPtr
		// use packet
		_ = pkt
		pool.Put(pktPtr)
	}
}
