package wire

import (
	"encoding/binary"
	"errors"
)

const (
	// HeaderSize is 96 bytes: 32+32 (IDs) + 4 (FlowLabel) + 8 (SeqNum) + 16 (SessionID) + 2 (Flags) + 2 (Length)
	HeaderSize = 96
)

var (
	ErrBufferTooSmall = errors.New("wire: buffer too small")
	ErrInvalidHeader  = errors.New("wire: invalid header")
)

// Header is the SMIP/MWP sovereign wire format header.
// All fields are big-endian on the wire.
type Header struct {
	SrcID     [32]byte // Sovereign Crypto ID of sender
	DstID     [32]byte // Sovereign Crypto ID of destination
	FlowLabel uint32
	SeqNum    uint64
	SessionID [16]byte // PQC ephemeral session reference
	Flags     uint16   // Mode bits: datagram/reliable, attestation present, etc.
	Length    uint16   // Payload length (excluding header)
}

// Flag bit definitions
const (
	FlagDatagram    uint16 = 1 << 0 // Unreliable datagram mode
	FlagReliable    uint16 = 1 << 1 // Reliable stream mode
	FlagAttestation uint16 = 1 << 2 // Attestation tag present after header
	FlagPQC         uint16 = 1 << 3 // PQC-hybrid session
	FlagEncrypted   uint16 = 1 << 4 // Payload is encrypted
)

// Marshal serialises the header into buf (must be >= HeaderSize bytes).
func (h *Header) Marshal(buf []byte) error {
	if len(buf) < HeaderSize {
		return ErrBufferTooSmall
	}
	copy(buf[0:32], h.SrcID[:])
	copy(buf[32:64], h.DstID[:])
	binary.BigEndian.PutUint32(buf[64:68], h.FlowLabel)
	binary.BigEndian.PutUint64(buf[68:76], h.SeqNum)
	copy(buf[76:92], h.SessionID[:])
	binary.BigEndian.PutUint16(buf[92:94], h.Flags)
	binary.BigEndian.PutUint16(buf[94:96], h.Length)
	return nil
}

// Unmarshal deserialises a header from buf.
func (h *Header) Unmarshal(buf []byte) error {
	if len(buf) < HeaderSize {
		return ErrBufferTooSmall
	}
	copy(h.SrcID[:], buf[0:32])
	copy(h.DstID[:], buf[32:64])
	h.FlowLabel = binary.BigEndian.Uint32(buf[64:68])
	h.SeqNum = binary.BigEndian.Uint64(buf[68:76])
	copy(h.SessionID[:], buf[76:92])
	h.Flags = binary.BigEndian.Uint16(buf[92:94])
	h.Length = binary.BigEndian.Uint16(buf[94:96])
	return nil
}

// ParseHeader is a convenience wrapper that returns a populated Header.
func ParseHeader(buf []byte) (Header, error) {
	var h Header
	return h, h.Unmarshal(buf)
}

// IsDatagram reports whether the datagram flag is set.
func (h *Header) IsDatagram() bool { return h.Flags&FlagDatagram != 0 }

// IsEncrypted reports whether the encrypted flag is set.
func (h *Header) IsEncrypted() bool { return h.Flags&FlagEncrypted != 0 }

// SetEncrypted sets the encrypted flag.
func (h *Header) SetEncrypted(v bool) {
	if v {
		h.Flags |= FlagEncrypted
	} else {
		h.Flags &^= FlagEncrypted
	}
}
