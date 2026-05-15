package afxdp

import (
    "context"
    stdRand "crypto/rand"
    "testing"
    "time"

    "smip-mwp/internal/crypto"
    "smip-mwp/internal/routing"
    "smip-mwp/internal/wire"
)

func TestRunXDPLoop_InPlaceEncryptDecrypt(t *testing.T) {
    fwd := &Forwarder{routeTable: routing.NewTable()}

    var src, dst [32]byte
    copy(src[:], []byte("inplace-src-000000000000000000000000"))
    copy(dst[:], []byte("inplace-dst-000000000000000000000000"))

    // Create a session and register it under a random session ID
    combined := make([]byte, 64)
    if _, err := stdRand.Read(combined); err != nil {
        t.Fatalf("rand: %v", err)
    }
    sessionInfo := append(src[:], dst[:]...)
    sess, err := crypto.NewHybridSession(combined, sessionInfo)
    if err != nil {
        t.Fatalf("new hybrid session: %v", err)
    }

    var sid [16]byte
    copy(sid[:], []byte("session-id-inplace"))
    fwd.AddSession(sid, &Session{CryptoState: sess})

    // Prime a route for the destination
    var next [32]byte
    copy(next[:], []byte("next-hop-inplace-0000000000000000"))
    fwd.routeTable.UpdateRoute(routing.RouteEntry{DestID: dst, NextHopID: next})

    // Create a header with payload and extra capacity for tag
    payload := []byte("hello zero-copy world")
    h := wire.Header{SrcID: src, DstID: dst, FlowLabel: 0x2, SeqNum: 42, SessionID: sid, Length: uint16(len(payload))}
    buf := make([]byte, wire.HeaderSize+len(payload), wire.HeaderSize+len(payload)+crypto.TagSize)
    if err := h.Marshal(buf); err != nil {
        t.Fatalf("marshal header: %v", err)
    }
    copy(buf[wire.HeaderSize:], payload)

    sock := newTestSocket()
    umem := &testUMEM{}

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    go fwd.RunXDPLoop(ctx, sock, umem)

    // send buffer
    sock.frames <- buf

    // expect sent packet
    select {
    case sent := <-sock.sent:
        if len(sent) != 1 {
            t.Fatalf("expected 1 sent pkt, got %d", len(sent))
        }
        pkt := sent[0]
        // parse header
        ph, err := wire.ParseHeader(pkt)
        if err != nil {
            t.Fatalf("parse sent header: %v", err)
        }
        if int(ph.Length) != len(payload)+crypto.TagSize {
            t.Fatalf("expected payload length %d, got %d", len(payload)+crypto.TagSize, ph.Length)
        }
        // attempt decrypt using session
        pt, err := sess.DecryptInPlace(pkt[wire.HeaderSize:wire.HeaderSize+int(ph.Length)], ph.SeqNum)
        if err != nil {
            t.Fatalf("decrypt inplace failed: %v", err)
        }
        if string(pt) != string(payload) {
            t.Fatalf("plaintext mismatch: got %q want %q", string(pt), string(payload))
        }
    case <-time.After(1 * time.Second):
        t.Fatalf("timed out waiting for sent pkt")
    }
}
