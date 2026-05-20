// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright 2026 rwilliamspbg-ops
//
// This file is part of SMIP-MWP.
// SMIP-MWP is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation; either version 3 of the License, or (at your option) any later version.
// See the LICENSE file in the project root for details.

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
	pollT  *time.Timer
	pollB  [][]byte
	// lastSent holds the most recent packet batch for inspection by tests.
	lastSent   [][]byte
	sentSignal chan struct{}
}

func newTestSocket() *testSocket {
	t := time.NewTimer(200 * time.Millisecond)
	if !t.Stop() {
		select {
		case <-t.C:
		default:
		}
	}
	return &testSocket{frames: make(chan []byte, 4), sent: make(chan [][]byte, 4), pollT: t, pollB: make([][]byte, 1), sentSignal: make(chan struct{}, 1)}
}

func (s *testSocket) Poll(max int) ([][]byte, error) {
	s.mu.Lock()
	s.pollT.Reset(200 * time.Millisecond)
	s.mu.Unlock()

	select {
	case b := <-s.frames:
		s.mu.Lock()
		if !s.pollT.Stop() {
			select {
			case <-s.pollT.C:
			default:
			}
		}
		s.pollB[0] = b
		s.mu.Unlock()
		return s.pollB, nil
	case <-s.pollT.C:
		return nil, nil
	}
}

func (s *testSocket) Send(pkts [][]byte) error {
	s.mu.Lock()
	// record last sent batch for test inspection
	s.lastSent = pkts
	s.mu.Unlock()
	// notify any waiter without blocking (benchmarks use sentSignal)
	select {
	case s.sentSignal <- struct{}{}:
	default:
	}
	// also deliver on the older sent channel for tests that read it; do not block
	select {
	case s.sent <- pkts:
	default:
	}
	return nil
}

func (s *testSocket) Close() error {
	s.mu.Lock()
	if s.pollT != nil {
		s.pollT.Stop()
	}
	s.mu.Unlock()
	close(s.frames)
	close(s.sent)
	close(s.sentSignal)
	return nil
}

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
	case <-sock.sentSignal:
		sock.mu.Lock()
		sent := sock.lastSent
		sock.mu.Unlock()
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
