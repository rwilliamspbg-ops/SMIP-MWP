package afxdp

import (
    "context"
    "testing"
    "time"

    "smip-mwp/internal/routing"
    "smip-mwp/internal/wire"
)

func TestForwarderIntegration_stub(t *testing.T) {
    // Build a header and ensure parsing + routing + forwarder lifecycle work together.
    var src, dst [32]byte
    copy(src[:], []byte("src-integration-test-00000000000000"))
    copy(dst[:], []byte("dst-integration-test-00000000000000"))

    h := wire.Header{
        SrcID:     src,
        DstID:     dst,
        FlowLabel: 0x1234abcd,
        SeqNum:    1,
        Flags:     0x1,
        Length:    0,
    }

    buf := wire.NewHeaderBuffer(int(h.Length))
    if err := h.Marshal(buf); err != nil {
        t.Fatalf("marshal: %v", err)
    }

    parsed, err := wire.ParseHeader(buf)
    if err != nil {
        t.Fatalf("parse: %v", err)
    }
    if parsed.FlowLabel != h.FlowLabel {
        t.Fatalf("flowlabel mismatch")
    }

    // Prepare routing table with a route for dst
    rt := routing.NewTable()
    var next [32]byte
    copy(next[:], []byte("next-hop-integration-test-0000000000"))
    rt.UpdateRoute(routing.RouteEntry{DestID: dst, NextHopID: next})

    // Create forwarder stub and run for a short period to exercise Run/Close.
    cfg := Config{Interface: "lo", QueueID: 0, ZeroCopy: false, NumFrames: 64, FrameSize: 2048, BatchSize: 32}
    fwd, err := NewForwarder(cfg, rt)
    if err != nil {
        t.Fatalf("new forwarder: %v", err)
    }

    ctx, cancel := context.WithCancel(context.Background())
    go fwd.Run(ctx)

    // let it run a couple ticks
    time.Sleep(250 * time.Millisecond)
    cancel()

    // allow graceful shutdown
    time.Sleep(50 * time.Millisecond)
    if err := fwd.Close(); err != nil {
        t.Fatalf("close: %v", err)
    }
}
