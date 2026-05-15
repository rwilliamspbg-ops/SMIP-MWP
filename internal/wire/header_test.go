package wire

import (
	"bytes"
	"crypto/rand"
	"testing"
)

func random32(t *testing.T) [32]byte {
	var b [32]byte
	if _, err := rand.Read(b[:]); err != nil {
		t.Fatalf("rand: %v", err)
	}
	return b
}

func TestMarshalAndParse(t *testing.T) {
	src := random32(t)
	dst := random32(t)
	var sid [16]byte
	if _, err := rand.Read(sid[:]); err != nil {
		t.Fatalf("rand sid: %v", err)
	}

	h := Header{
		SrcID:     src,
		DstID:     dst,
		FlowLabel: 0xdeadbeef,
		SeqNum:    42,
		SessionID: sid,
		Flags:     0x1,
		Length:    128,
	}

	buf := NewHeaderBuffer(int(h.Length))
	if err := h.Marshal(buf); err != nil {
		t.Fatalf("marshal: %v", err)
	}

	parsed, err := ParseHeader(buf)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	if !bytes.Equal(parsed.SrcID[:], h.SrcID[:]) {
		t.Fatalf("src mismatch")
	}
	if !bytes.Equal(parsed.DstID[:], h.DstID[:]) {
		t.Fatalf("dst mismatch")
	}
	if parsed.FlowLabel != h.FlowLabel || parsed.SeqNum != h.SeqNum || parsed.Length != h.Length {
		t.Fatalf("fields mismatch")
	}
}
