//go:build withafxdp
// +build withafxdp

package afxdp

// XDPDescriptor represents a single packet descriptor in UMEM.
// This is a simplified representation; real AF_XDP would use xdp.Descriptor.
type XDPDescriptor struct {
	Addr uint64 // Physical address in UMEM
	Len  uint32 // Packet length
	Pad  uint32 // Padding for alignment
}

// xdpSocketImpl provides a mock AF_XDP socket implementation
// In production, this would wrap github.com/asavie/xdp.Socket
type xdpSocketImpl struct {
	descriptors []*XDPDescriptor
	fillIdx     int
	rxIdx       int
	txIdx       int
	compIdx     int
}

// NumFreeFillSlots returns the number of free slots in the fill ring
func (s *xdpSocketImpl) NumFreeFillSlots() int {
	return 128 - s.fillIdx // Mock implementation
}

// GetDescs returns available descriptors from the UMEM pool
func (s *xdpSocketImpl) GetDescs(n int) []*XDPDescriptor {
	descs := make([]*XDPDescriptor, 0, n)
	for i := 0; i < n && i < 128; i++ {
		descs = append(descs, &XDPDescriptor{
			Addr: uint64(i * 2048), // Assume 2KB frames
			Len:  2048,
		})
	}
	return descs
}

// Fill pushes descriptors to the fill ring (marks UMEM as available for RX)
func (s *xdpSocketImpl) Fill(descs []*XDPDescriptor) {
	s.fillIdx = (s.fillIdx + len(descs)) % 128
}

// Poll polls for RX and completion events, returning (numRx, numCompleted, err)
func (s *xdpSocketImpl) Poll(maxEvents int) (int, int, error) {
	// Mock: return 0 RX (no packets), 0 completed
	return 0, 0, nil
}

// Receive returns received packet descriptors
func (s *xdpSocketImpl) Receive(n int) []*XDPDescriptor {
	return make([]*XDPDescriptor, 0)
}

// GetFrame returns the actual packet data for a descriptor
func (s *xdpSocketImpl) GetFrame(d *XDPDescriptor) []byte {
	if d == nil {
		return nil
	}
	// Mock: return empty frame
	return make([]byte, d.Len)
}

// Complete marks transmitted descriptors as completed (returns to UMEM)
func (s *xdpSocketImpl) Complete(n int) {
	s.compIdx = (s.compIdx + n) % 128
}

// Transmit sends packet descriptors to TX ring
func (s *xdpSocketImpl) Transmit(descs []*XDPDescriptor) {
	s.txIdx = (s.txIdx + len(descs)) % 128
}

// Close releases socket resources
func (s *xdpSocketImpl) Close() error {
	return nil
}
