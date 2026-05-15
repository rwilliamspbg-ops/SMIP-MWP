//go:build withafxdp
// +build withafxdp

package afxdp

import (
	"fmt"
	"time"

	xdp "github.com/asavie/xdp"
)

// XDPSocket wraps an asavie/xdp socket bound to a network interface and queue.
type XDPSocket struct {
	s *xdp.Socket
}

// NewXDPSocket creates and binds an AF_XDP socket to iface:queue using the
// provided UMEM. This implementation uses the asavie/xdp library; adjust the
// call sites if the library API changes.
func NewXDPSocket(iface string, queue int, umem *UMEM) (*XDPSocket, error) {
	if iface == "" {
		return nil, fmt.Errorf("iface required")
	}
	if umem == nil || umem.u == nil {
		return nil, fmt.Errorf("umem required")
	}

	// Create socket config
	cfg := xdp.SocketConfig{
		Iface: iface,
		Queue: uint32(queue),
	}
	sock, err := xdp.NewSocket(&cfg, umem.u)
	if err != nil {
		return nil, fmt.Errorf("xdp.NewSocket: %w", err)
	}
	return &XDPSocket{s: sock}, nil
}

// Close releases socket resources.
func (s *XDPSocket) Close() error {
	if s == nil || s.s == nil {
		return nil
	}
	return s.s.Close()
}

// Poll attempts to receive up to `max` frames from the underlying xdp.Socket.
// This method uses reflection to call the available batch/recv API on the
// asavie/xdp Socket implementation to remain compatible with minor API
// variations between versions. It returns a slice of byte slices (one per
// received frame).
func (s *XDPSocket) Poll(max int) ([][]byte, error) {
	if s == nil || s.s == nil {
		return nil, fmt.Errorf("socket not initialized")
	}
	xsk := s.s

	// Refill UMEM descriptors
	free := xsk.NumFreeFillSlots()
	if free > 0 {
		descs := xsk.GetDescs(free)
		if len(descs) > 0 {
			xsk.Fill(descs)
		}
	}

	numRx, _, err := xsk.Poll(0)
	if err != nil {
		return nil, fmt.Errorf("xdp poll: %w", err)
	}
	if numRx == 0 {
		return nil, nil
	}
	descs := xsk.Receive(numRx)
	out := make([][]byte, 0, len(descs))
	for _, d := range descs {
		out = append(out, xsk.GetFrame(d))
	}
	return out, nil
}

// Send attempts to transmit a batch of packets using the underlying xdp.Socket.
// It uses reflection to call common transmit APIs (SendBatch, Transmit, Write,
// Tx) on the Socket. It returns an error if no compatible API is found.
func (s *XDPSocket) Send(pkts [][]byte) error {
	if s == nil || s.s == nil {
		return fmt.Errorf("socket not initialized")
	}
	xsk := s.s
	total := len(pkts)
	idx := 0
	for idx < total {
		avail := xsk.NumFreeTxSlots()
		if avail == 0 {
			// poll briefly to drive completions
			_, _, err := xsk.Poll(1)
			if err != nil {
				time.Sleep(1 * time.Millisecond)
				continue
			}
			avail = xsk.NumFreeTxSlots()
			if avail == 0 {
				time.Sleep(1 * time.Millisecond)
				continue
			}
		}

		n := avail
		remaining := total - idx
		if n > remaining {
			n = remaining
		}

		descs := xsk.GetDescs(n)
		if len(descs) == 0 {
			time.Sleep(1 * time.Millisecond)
			continue
		}

		for i := 0; i < len(descs); i++ {
			d := descs[i]
			frame := xsk.GetFrame(d)
			pkt := pkts[idx+i]
			if len(pkt) > int(d.Len) {
				return fmt.Errorf("packet too large for frame: %d > %d", len(pkt), d.Len)
			}
			copy(frame, pkt)
			descs[i].Len = uint32(len(pkt))
		}

		xsk.Transmit(descs)
		idx += len(descs)
	}

	if numCompleted := xsk.NumCompleted(); numCompleted > 0 {
		xsk.Complete(numCompleted)
	}
	return nil
}
