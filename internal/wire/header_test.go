package wire

import (
    "bytes"
    "testing"
)

func TestMarshalParseRoundTrip(t *testing.T) {
    var h Header
    copy(h.SrcID[:], []byte("src-id-0123456789012345678901"))
    copy(h.DstID[:], []byte("dst-id-0123456789012345678901"))
    h.FlowLabel = 0xdeadbeef
    h.SeqNum = 0x1122334455667788
    copy(h.SessionID[:], []byte("session-01234567"))
    h.Flags = 0xabba
    h.Length = 1234

    buf := NewHeaderBuffer(0)
    if err := h.Marshal(buf[:HeaderSize]); err != nil {
        t.Fatalf("Marshal failed: %v", err)
    }

    got, err := ParseHeader(buf)
    if err != nil {
        t.Fatalf("ParseHeader failed: %v", err)
    }

    if !bytes.Equal(got.SrcID[:], h.SrcID[:]) {
        t.Fatalf("SrcID mismatch")
    }
    if !bytes.Equal(got.DstID[:], h.DstID[:]) {
        t.Fatalf("DstID mismatch")
    }
    if got.FlowLabel != h.FlowLabel || got.SeqNum != h.SeqNum || got.Flags != h.Flags || got.Length != h.Length {
        t.Fatalf("Fields mismatch: got=%+v want=%+v", got, h)
    }
}

func TestZeroCopySetters(t *testing.T) {
    buf := NewHeaderBuffer(0)
    v, err := ViewHeader(buf)
    if err != nil {
        t.Fatalf("ViewHeader: %v", err)
    }

    var src [32]byte
    copy(src[:], []byte("zero-copy-src-id-0123456789"))
    v.SetSrcID(src)
    var dst [32]byte
    copy(dst[:], []byte("zero-copy-dst-id-0123456789"))
    v.SetDstID(dst)
    v.SetFlowLabel(0xabadcafe)
    v.SetSeqNum(0x1020304050607080)
    var sid [16]byte
    copy(sid[:], []byte("sess-0000000000"))
    v.SetSessionID(sid)
    v.SetFlags(0xface)
    v.SetLength(42)

    parsed, err := ParseHeader(buf)
    if err != nil {
        t.Fatalf("ParseHeader: %v", err)
    }
    if parsed.FlowLabel != 0xabadcafe || parsed.SeqNum != 0x1020304050607080 || parsed.Flags != 0xface || parsed.Length != 42 {
        t.Fatalf("Parsed values mismatch: %+v", parsed)
    }
}


