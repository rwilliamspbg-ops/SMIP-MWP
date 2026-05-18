//go:build withafxdp && asavie
// +build withafxdp,asavie

package afxdp

import (
	"fmt"
	"sync"

	_ "github.com/asavie/xdp"
)

// NOTE: This file is a scaffold for integrating the real github.com/asavie/xdp
// library. It is only built when both build tags `withafxdp` and `asavie`
// are provided. Implementations below should be replaced with calls into the
// asavie/xdp API to allocate kernel UMEM, pin pages, and expose frame indices.

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

// NewUMEM creates a UMEM region backed by the kernel using asavie/xdp.
// Replace the body with real asavie/xdp calls (e.g., allocate, configure,
// and return a wrapper that implements the UMEM methods used by forwarder).
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
	// Preallocate frames to match asavie expectations (frame mapping will
	// be managed by xdp.Socket when binding; preallocating avoids nil frames)
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
