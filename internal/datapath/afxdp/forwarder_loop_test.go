package afxdp

import (
	"context"
	"reflect"
	"sync"
	"testing"
	"time"

	"smip-mwp/internal/routing"
	"smip-mwp/internal/wire"
)

type testSocket struct {
	frames chan []byte
	sent   chan [][]byte
	mu     sync.Mutex
}

func newTestSocket() *testSocket {
	return &testSocket{frames: make(chan []byte, 4), sent: make(chan [][]byte, 4)}
}

func (s *testSocket) Poll(max int) ([][]byte, error) {
	select {
	case b := <-s.frames:
		return [][]byte{b}, nil
	case <-time.After(200 * time.Millisecond):
		return nil, nil
	}
}

func (s *testSocket) Send(pkts [][]byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sent <- pkts
	return nil
}

func (s *testSocket) Close() error { close(s.frames); close(s.sent); return nil }

type testUMEM struct{ closed bool }

func (u *testUMEM) Close() error { u.closed = true; return nil }

func TestRunXDPLoop_SendsReceivedPackets(t *testing.T) {
	fwd := &Forwarder{routeTable: routing.NewTable(), logger: nil}

	// Set a route for the destination used in the test packet
	var src, dst [32]byte
	copy(src[:], []byte("runloop-src-00000000000000000000000"))
	copy(dst[:], []byte("runloop-dst-0000000000000000000000000"))
	var next [32]byte
	copy(next[:], []byte("runloop-next-00000000000000000000000"))
	fwd.routeTable.UpdateRoute(routing.RouteEntry{DestID: dst, NextHopID: next})

	// Create a header and frame buffer
	h := wire.Header{SrcID: src, DstID: dst, FlowLabel: 0x1, Length: 0}
	buf := wire.NewHeaderBuffer(int(h.Length))
	if err := h.Marshal(buf); err != nil {
		t.Fatalf("marshal: %v", err)
	}

	sock := newTestSocket()
	umem := &testUMEM{}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Run loop
	go fwd.RunXDPLoop(ctx, sock, umem)

	// Send a frame into the socket
	sock.frames <- buf

	// Expect a send within a short time
	select {
	case sent := <-sock.sent:
		if len(sent) != 1 {
			t.Fatalf("unexpected sent count: %d", len(sent))
		}
		if !reflect.DeepEqual(sent[0], buf) {
			t.Fatalf("sent buffer mismatch")
		}
	case <-time.After(1 * time.Second):
		t.Fatalf("timed out waiting for send")
	}

	cancel()
	// Give loop time to exit
	time.Sleep(10 * time.Millisecond)
}
