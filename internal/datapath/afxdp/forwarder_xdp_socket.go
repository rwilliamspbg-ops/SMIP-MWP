//go:build withafxdp
// +build withafxdp

package afxdp

import (
	"fmt"
)

// XDPSocket wraps low-level AF_XDP socket operations.
// This implementation provides a high-level abstraction that can be swapped
// for real AF_XDP implementation when library APIs are available.
type XDPSocket struct {
	s *xdpSocketImpl // Mock implementation for now
}

// NewXDPSocket creates an AF_XDP socket bound to iface:queue.
// In production, this would use github.com/asavie/xdp library.
func NewXDPSocket(iface string, queue int, umem *UMEM) (*XDPSocket, error) {
	if iface == "" {
		return nil, fmt.Errorf("iface required")
	}
	if umem == nil {
		return nil, fmt.Errorf("umem required")
	}

	// Create mock socket for now (can be replaced with real AF_XDP later)
	sock := &XDPSocket{
		s: &xdpSocketImpl{
			descriptors: make([]*XDPDescriptor, 0),
			fillIdx:     0,
			rxIdx:       0,
			txIdx:       0,
			compIdx:     0,
		},
	}
	return sock, nil
}

// Close releases socket resources
func (s *XDPSocket) Close() error {
	if s == nil || s.s == nil {
		return nil
	}
	return s.s.Close()
}
