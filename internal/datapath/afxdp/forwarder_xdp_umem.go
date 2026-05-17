//go:build withafxdp
// +build withafxdp

package afxdp

import (
	"fmt"
)

// UMEM provides a mock User Memory region for AF_XDP.
// In production, this would use github.com/asavie/xdp UMEM implementation.
type UMEM struct {
	frames    [][]byte // Pool of frame buffers
	frameSize int
	numFrames int
}

// NewUMEM initializes a UMEM region as a mock implementation.
// In production, this would allocate actual kernel-managed memory via AF_XDP.
func NewUMEM(numFrames, frameSize int) (*UMEM, error) {
	if numFrames <= 0 || frameSize <= 0 {
		return nil, fmt.Errorf("invalid UMEM params: numFrames=%d frameSize=%d", numFrames, frameSize)
	}

	u := &UMEM{
		frames:    make([][]byte, numFrames),
		frameSize: frameSize,
		numFrames: numFrames,
	}

	// Preallocate frame buffers
	for i := 0; i < numFrames; i++ {
		u.frames[i] = make([]byte, frameSize)
	}

	return u, nil
}

// Close releases UMEM resources
func (u *UMEM) Close() error {
	if u == nil {
		return nil
	}
	u.frames = nil
	return nil
}
