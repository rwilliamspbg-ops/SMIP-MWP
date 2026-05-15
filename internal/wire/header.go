package wire

import (
	"encoding/binary"
	"errors"
)

const (
	// HeaderSize defines the fixed header length in bytes.
	HeaderSize = 96 // room for IDs + fields, adjustable

	srcOffset     = 0
	dstOffset     = 32
	flowOffset    = 64
	seqOffset     = 68
	sessionOffset = 76
	flagsOffset   = 92
	lenOffset     = 94
)

var (
	ErrBufferTooSmall = errors.New("wire: buffer too small for header")
)

// HeaderView provides a zero-allocation view over a header-backed byte slice.
type HeaderView struct {
	buf []byte
}

// NewHeaderBuffer returns a byte slice pre-sized for a header plus payload of given length.
func NewHeaderBuffer(payloadLen int) []byte {
	b := make([]byte, HeaderSize+payloadLen)
	return b
}

// ViewHeader returns a HeaderView over buf; buffer must be at least HeaderSize bytes.
func ViewHeader(buf []byte) (*HeaderView, error) {
	if len(buf) < HeaderSize {
		return nil, ErrBufferTooSmall
	}
	return &HeaderView{buf: buf}, nil
}

// ParseHeader parses the header into a copy-structured Header.
func ParseHeader(buf []byte) (Header, error) {
	v, err := ViewHeader(buf)
	if err != nil {
		return Header{}, err
	}
	var h Header
	copy(h.SrcID[:], v.SrcID())
	copy(h.DstID[:], v.DstID())
	h.FlowLabel = v.FlowLabel()
	h.SeqNum = v.SeqNum()
	copy(h.SessionID[:], v.SessionID())
	h.Flags = v.Flags()
	h.Length = v.Length()
	return h, nil
}

// Header is the copy-structured header (for convenience when building new headers).
type Header struct {
	SrcID     [32]byte
	DstID     [32]byte
	FlowLabel uint32
	SeqNum    uint64
	SessionID [16]byte
	Flags     uint16
	Length    uint16
}

// Marshal writes Header into buf (must be at least HeaderSize bytes).
func (h *Header) Marshal(buf []byte) error {
	if len(buf) < HeaderSize {
		return ErrBufferTooSmall
	}
	copy(buf[srcOffset:srcOffset+32], h.SrcID[:])
	copy(buf[dstOffset:dstOffset+32], h.DstID[:])
	binary.BigEndian.PutUint32(buf[flowOffset:flowOffset+4], h.FlowLabel)
	binary.BigEndian.PutUint64(buf[seqOffset:seqOffset+8], h.SeqNum)
	copy(buf[sessionOffset:sessionOffset+16], h.SessionID[:])
	binary.BigEndian.PutUint16(buf[flagsOffset:flagsOffset+2], h.Flags)
	binary.BigEndian.PutUint16(buf[lenOffset:lenOffset+2], h.Length)
	return nil
}

// --- HeaderView accessors (zero-copy) ---
func (v *HeaderView) SrcID() []byte { return v.buf[srcOffset : srcOffset+32] }
func (v *HeaderView) DstID() []byte { return v.buf[dstOffset : dstOffset+32] }
func (v *HeaderView) FlowLabel() uint32 {
	return binary.BigEndian.Uint32(v.buf[flowOffset : flowOffset+4])
}
func (v *HeaderView) SeqNum() uint64    { return binary.BigEndian.Uint64(v.buf[seqOffset : seqOffset+8]) }
func (v *HeaderView) SessionID() []byte { return v.buf[sessionOffset : sessionOffset+16] }
func (v *HeaderView) Flags() uint16 {
	return binary.BigEndian.Uint16(v.buf[flagsOffset : flagsOffset+2])
}
func (v *HeaderView) Length() uint16 { return binary.BigEndian.Uint16(v.buf[lenOffset : lenOffset+2]) }

// Setters
func (v *HeaderView) SetSrcID(id [32]byte) { copy(v.SrcID(), id[:]) }
func (v *HeaderView) SetDstID(id [32]byte) { copy(v.DstID(), id[:]) }
func (v *HeaderView) SetFlowLabel(l uint32) {
	binary.BigEndian.PutUint32(v.buf[flowOffset:flowOffset+4], l)
}
func (v *HeaderView) SetSeqNum(n uint64)       { binary.BigEndian.PutUint64(v.buf[seqOffset:seqOffset+8], n) }
func (v *HeaderView) SetSessionID(id [16]byte) { copy(v.SessionID(), id[:]) }
func (v *HeaderView) SetFlags(f uint16) {
	binary.BigEndian.PutUint16(v.buf[flagsOffset:flagsOffset+2], f)
}
func (v *HeaderView) SetLength(l uint16) { binary.BigEndian.PutUint16(v.buf[lenOffset:lenOffset+2], l) }
