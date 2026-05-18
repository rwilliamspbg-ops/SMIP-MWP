//go:build withafxdp && !asavie
// +build withafxdp,!asavie

package afxdp

import (
	"fmt"
)

// XDPSocket wraps low-level AF_XDP socket operations.
// This implementation provides a high-level abstraction that can be swapped
// for real AF_XDP implementation when library APIs are available.
type XDPSocket struct {
	s socketBackend
}

// socketBackend abstracts the backing socket implementation used by XDPSocket.
// Both the mock `xdpSocketImpl` and the real `realSocketImpl` implement this.
type socketBackend interface {
	NumFreeFillSlots() int
	GetDescs(n int) []*XDPDescriptor
	Fill(descs []*XDPDescriptor)
	Poll(maxEvents int) (int, int, error)
	Receive(n int) []*XDPDescriptor
	GetFrame(d *XDPDescriptor) []byte
	Complete(n int)
	Transmit(descs []*XDPDescriptor)
	Close() error
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

	// Create mock socket backed by provided UMEM (simulates descriptors referencing UMEM frames)
	sock := &XDPSocket{
		s: &xdpSocketImpl{
			umem:     umem,
			fillRing: make([]*XDPDescriptor, 0),
			rxRing:   make([]*XDPDescriptor, 0),
			txRing:   make([]*XDPDescriptor, 0),
			compRing: make([]*XDPDescriptor, 0),
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
