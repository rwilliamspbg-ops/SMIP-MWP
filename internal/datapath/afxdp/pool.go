// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright 2026 rwilliamspbg-ops
// Frame pooling for AF_XDP zero-copy forwarding

package afxdp

import (
	"sync"
)

// FramePool provides a sync.Pool-based frame buffer pool to eliminate per-packet allocations
type FramePool struct {
	pool      sync.Pool
	size      int // Fixed size for all frames in pool
}

func NewFramePool(size int) *FramePool {
	return &FramePool{
		pool: sync.Pool{
			New: func() interface{} {
				buf := make([]byte, size)
				return (*[]byte)(&buf)
			},
		},
		size: size,
	}
}

// Get retrieves a frame buffer from the pool (or allocates if empty)
func (p *FramePool) Get() []byte {
	b := p.pool.Get().(*[]byte)
	*b = (*b)[:0] // Reset length, preserve capacity
	return *b
}

// Put returns a frame buffer to the pool for reuse
func (p *FramePool) Put(b []byte) {
	if cap(*b) != p.size {
		// Wrong size - discard to avoid pool poisoning
		return
	}
	p.pool.Put(b)
}

// GetWithLen retrieves a frame with specific length (falls back to allocation if wrong size)
func (p *FramePool) GetWithLen(size int) []byte {
	if size == p.size {
		b := p.pool.Get().(*[]byte)
		*b = (*b)[:0]
		return *b
	}
	// Fall back to allocation for non-standard sizes
	buf := make([]byte, size)
	return buf
}
