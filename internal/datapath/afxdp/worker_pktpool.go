// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright 2026 rwilliamspbg-ops
//
// This file is part of SMIP-MWP.
// SMIP-MWP is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation; either version 3 of the License, or (at your option) any later version.
// See the LICENSE file in the project root for details.

package afxdp

// workerPktPool is a simple single-consumer packet buffer pool intended to
// be used by a single worker goroutine. It avoids global sync.Pool contention
// by keeping buffers local to the worker.
type workerPktPool struct {
	bufs [][]byte
	size int
}

// newWorkerPktPool creates a pool with optional preallocation capacity.
func newWorkerPktPool(bufSize int, capacity int) *workerPktPool {
	p := &workerPktPool{size: bufSize}
	if capacity > 0 {
		p.bufs = make([][]byte, 0, capacity)
		for i := 0; i < capacity; i++ {
			p.bufs = append(p.bufs, make([]byte, bufSize))
		}
	}
	return p
}

// Get returns a pointer to a byte slice buffer. Caller should treat the
// buffer length as needed; pool guarantees capacity is at least bufSize.
func (p *workerPktPool) Get() *[]byte {
	if p == nil {
		b := make([]byte, p.size)
		return &b
	}
	n := len(p.bufs)
	if n == 0 {
		b := make([]byte, p.size)
		return &b
	}
	// pop
	buf := p.bufs[n-1]
	p.bufs = p.bufs[:n-1]
	return &buf
}

// Put returns a buffer to the pool.
func (p *workerPktPool) Put(b *[]byte) {
	if p == nil || b == nil {
		return
	}
	// reset length to zero but keep capacity
	*b = (*b)[:cap(*b)]
	p.bufs = append(p.bufs, *b)
}
