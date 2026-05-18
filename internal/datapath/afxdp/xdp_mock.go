//go:build withafxdp && !asavie
// +build withafxdp,!asavie

package afxdp

// XDPDescriptor represents a single packet descriptor in UMEM.
// This is a simplified representation; real AF_XDP would use xdp.Descriptor.
// XDPDescriptor represents a single packet descriptor in UMEM.
// This is a simplified representation; real AF_XDP would use xdp.Descriptor.

// xdpSocketImpl provides a mock AF_XDP socket implementation
// In production, this would wrap github.com/asavie/xdp.Socket
type xdpSocketImpl struct {
	umem *UMEM

	// simple rings represented as slices for the mock implementation
	fillRing []*XDPDescriptor // descriptors available for RX to fill into
	rxRing   []*XDPDescriptor // descriptors with received packets
	txRing   []*XDPDescriptor // descriptors pending TX completion
	compRing []*XDPDescriptor // descriptors completed and awaiting return
}

// NumFreeFillSlots returns the number of free slots in the fill ring
func (s *xdpSocketImpl) NumFreeFillSlots() int {
	if s == nil || s.umem == nil {
		return 0
	}
	// free slots limited by available UMEM frames
	s.umem.mu.Lock()
	free := len(s.umem.freeList)
	s.umem.mu.Unlock()
	return free
}

// GetDescs returns available descriptors from the UMEM pool
func (s *xdpSocketImpl) GetDescs(n int) []*XDPDescriptor {
	if s == nil || s.umem == nil || n <= 0 {
		return nil
	}
	idxs := s.umem.AllocateIndices(n)
	if len(idxs) == 0 {
		return nil
	}
	descs := make([]*XDPDescriptor, 0, len(idxs))
	for _, idx := range idxs {
		descs = append(descs, &XDPDescriptor{Addr: uint64(idx), Len: uint32(s.umem.frameSize)})
	}
	return descs
}

// Fill pushes descriptors to the fill ring (marks UMEM as available for RX)
func (s *xdpSocketImpl) Fill(descs []*XDPDescriptor) {
	if s == nil {
		return
	}
	if len(descs) == 0 {
		return
	}
	s.fillRing = append(s.fillRing, descs...)
}

// Poll polls for RX and completion events, returning (numRx, numCompleted, err)
func (s *xdpSocketImpl) Poll(maxEvents int) (int, int, error) {
	if s == nil {
		return 0, 0, nil
	}
	// Move any fillRing descriptors to rxRing to simulate packet reception
	// For the mock, we treat all filled descriptors as received immediately.
	if len(s.fillRing) > 0 {
		n := len(s.fillRing)
		if n > maxEvents {
			n = maxEvents
		}
		s.rxRing = append(s.rxRing, s.fillRing[:n]...)
		s.fillRing = s.fillRing[n:]
	}

	// Move txRing to compRing to simulate immediate completion
	if len(s.txRing) > 0 {
		s.compRing = append(s.compRing, s.txRing...)
		s.txRing = nil
	}

	return len(s.rxRing), len(s.compRing), nil
}

// Receive returns received packet descriptors
func (s *xdpSocketImpl) Receive(n int) []*XDPDescriptor {
	if s == nil || n <= 0 || len(s.rxRing) == 0 {
		return nil
	}
	if n > len(s.rxRing) {
		n = len(s.rxRing)
	}
	descs := s.rxRing[:n]
	s.rxRing = s.rxRing[n:]
	return descs
}

// GetFrame returns the actual packet data for a descriptor
func (s *xdpSocketImpl) GetFrame(d *XDPDescriptor) []byte {
	if s == nil || s.umem == nil || d == nil {
		return nil
	}
	idx := int(d.Addr)
	return s.umem.GetFrameByIndex(idx)
}

// Complete marks transmitted descriptors as completed (returns to UMEM)
func (s *xdpSocketImpl) Complete(n int) {
	if s == nil || s.umem == nil || n <= 0 || len(s.compRing) == 0 {
		return
	}
	// Return up to n descriptors' frame indices to UMEM free list
	if n > len(s.compRing) {
		n = len(s.compRing)
	}
	idxs := make([]int, 0, n)
	for i := 0; i < n; i++ {
		d := s.compRing[i]
		idxs = append(idxs, int(d.Addr))
	}
	s.compRing = s.compRing[n:]
	s.umem.ReturnIndices(idxs)
}

// Transmit sends packet descriptors to TX ring
func (s *xdpSocketImpl) Transmit(descs []*XDPDescriptor) {
	if s == nil || len(descs) == 0 {
		return
	}
	// Add to txRing; Poll will move them to compRing and Complete will return frames
	s.txRing = append(s.txRing, descs...)
}

// Close releases socket resources
func (s *xdpSocketImpl) Close() error {
	return nil
}
