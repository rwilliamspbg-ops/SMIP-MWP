//go:build withafxdp
// +build withafxdp

package afxdp

import (
	"fmt"

	xdp "github.com/asavie/xdp"
)

// UMEM wraps an asavie/xdp UMEM region. This file requires the `withafxdp`
// build tag because it depends on kernel-specific syscalls and system
// configuration.
type UMEM struct {
	u *xdp.Umem
}

// NewUMEM initializes a UMEM region via the asavie/xdp library. Note: the
// asavie/xdp API may evolve; when building with `-tags=withafxdp` run
// `go mod tidy` to fetch dependencies and adjust imports if needed.
func NewUMEM(numFrames, frameSize int) (*UMEM, error) {
	if numFrames <= 0 || frameSize <= 0 {
		return nil, fmt.Errorf("invalid UMEM params")
	}

	// The asavie/xdp library expects a configuration to allocate UMEM.
	cfg := xdp.UmemConfig{
		NumFrames: uint32(numFrames),
		FrameSize: uint32(frameSize),
	}
	u, err := xdp.NewUmem(&cfg)
	if err != nil {
		return nil, fmt.Errorf("xdp.NewUmem: %w", err)
	}
	return &UMEM{u: u}, nil
}

// Close releases UMEM resources.
func (u *UMEM) Close() error {
	if u == nil || u.u == nil {
		return nil
	}
	return u.u.Close()
}
