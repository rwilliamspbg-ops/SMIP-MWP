//go:build withafxdp && !asavie
// +build withafxdp,!asavie

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
	"sync"
)

// UMEM provides a simple user-space-backed UMEM region for AF_XDP tests.
// It manages preallocated frames and a free frame pool so descriptors can
// reuse buffers without per-packet allocations.
type UMEM struct {
	frames    [][]byte // frame buffers
	frameSize int
	numFrames int

	mu       sync.Mutex
	freeList []int // indices of free frames
}

// NewUMEM initializes a UMEM region backed by preallocated frames.
func NewUMEM(numFrames, frameSize int) (*UMEM, error) {
	if numFrames <= 0 || frameSize <= 0 {
		return nil, fmt.Errorf("invalid UMEM params: numFrames=%d frameSize=%d", numFrames, frameSize)
	}

	u := &UMEM{
		frames:    make([][]byte, numFrames),
		frameSize: frameSize,
		numFrames: numFrames,
		freeList:  make([]int, 0, numFrames),
	}

	// Preallocate frame buffers and push indices to free list
	for i := 0; i < numFrames; i++ {
		u.frames[i] = make([]byte, frameSize)
		u.freeList = append(u.freeList, i)
	}

	return u, nil
}

// Close releases UMEM resources
func (u *UMEM) Close() error {
	if u == nil {
		return nil
	}
	u.mu.Lock()
	defer u.mu.Unlock()
	u.frames = nil
	u.freeList = nil
	return nil
}

// AllocateIndices reserves up to n free frame indices and returns them.
// The returned indices can be used as descriptor Addr fields (uint64).
func (u *UMEM) AllocateIndices(n int) []int {
	u.mu.Lock()
	defer u.mu.Unlock()
	if n <= 0 || len(u.freeList) == 0 {
		return nil
	}
	if n > len(u.freeList) {
		n = len(u.freeList)
	}
	res := make([]int, n)
	copy(res, u.freeList[len(u.freeList)-n:])
	u.freeList = u.freeList[:len(u.freeList)-n]
	return res
}

// ReturnIndices returns frame indices back to the free list for reuse.
func (u *UMEM) ReturnIndices(idxs []int) {
	if len(idxs) == 0 {
		return
	}
	u.mu.Lock()
	defer u.mu.Unlock()
	u.freeList = append(u.freeList, idxs...)
}

// GetFrameByIndex returns the frame slice for the given index.
func (u *UMEM) GetFrameByIndex(idx int) []byte {
	u.mu.Lock()
	defer u.mu.Unlock()
	if idx < 0 || idx >= len(u.frames) {
		return nil
	}
	return u.frames[idx]
}
