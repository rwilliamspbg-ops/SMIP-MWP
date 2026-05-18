package afxdp

// XDPDescriptor represents a single packet descriptor in UMEM.
// Addr is the frame index (not byte offset) used by the forwarder to index
// into UMEM frame arrays. Len is the frame length in bytes.
type XDPDescriptor struct {
	Addr uint64 // frame index
	Len  uint32 // Packet length
	Pad  uint32 // Padding for alignment
}
