package wire

import (
	"bytes"
	"testing"
)

func TestHeaderRoundTrip(t *testing.T) {
	orig := Header{
		FlowLabel: 0xDEADBEEF,
		SeqNum:    42,
		Flags:     FlagPQC | FlagEncrypted,
		Length:    512,
	}
	for i := range orig.SrcID {
		orig.SrcID[i] = byte(i)
	}
	for i := range orig.DstID {
		orig.DstID[i] = byte(255 - i)
	}
	for i := range orig.SessionID {
		orig.SessionID[i] = byte(i * 2)
	}

	buf := make([]byte, HeaderSize)
	if err := orig.Marshal(buf); err != nil {
		t.Fatal(err)
	}

	var parsed Header
	if err := parsed.Unmarshal(buf); err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(orig.SrcID[:], parsed.SrcID[:]) {
		t.Error("SrcID mismatch")
	}
	if !bytes.Equal(orig.DstID[:], parsed.DstID[:]) {
		t.Error("DstID mismatch")
	}
	if orig.FlowLabel != parsed.FlowLabel {
		t.Errorf("FlowLabel: got %d want %d", parsed.FlowLabel, orig.FlowLabel)
	}
	if orig.SeqNum != parsed.SeqNum {
		t.Errorf("SeqNum: got %d want %d", parsed.SeqNum, orig.SeqNum)
	}
	if orig.Flags != parsed.Flags {
		t.Errorf("Flags: got %d want %d", parsed.Flags, orig.Flags)
	}
	if orig.Length != parsed.Length {
		t.Errorf("Length: got %d want %d", parsed.Length, orig.Length)
	}
}

func TestBufferTooSmall(t *testing.T) {
	var h Header
	if err := h.Marshal(make([]byte, HeaderSize-1)); err != ErrBufferTooSmall {
		t.Errorf("expected ErrBufferTooSmall, got %v", err)
	}
	if err := h.Unmarshal(make([]byte, HeaderSize-1)); err != ErrBufferTooSmall {
		t.Errorf("expected ErrBufferTooSmall, got %v", err)
	}
}

func BenchmarkMarshal(b *testing.B) {
	h := Header{FlowLabel: 1, SeqNum: 100, Flags: FlagPQC | FlagEncrypted, Length: 1400}
	buf := make([]byte, HeaderSize)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = h.Marshal(buf)
	}
}
