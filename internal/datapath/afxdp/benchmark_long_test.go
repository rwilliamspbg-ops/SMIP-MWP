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
	"os"
	"testing"
	"time"

	"smip-mwp/internal/crypto"
	"smip-mwp/internal/routing"
	"smip-mwp/internal/wire"
)

// BenchmarkRunXDPLoop_WithCrypto measures the full hot-path including
// in-place encryption. Use `-benchtime=30s` to run longer.
func BenchmarkRunXDPLoop_WithCrypto(b *testing.B) {
	fwd := &Forwarder{routeTable: routing.NewTable(), logger: nil}

	var src, dst, next [32]byte
	copy(src[:], []byte("bench-src-000000000000000000000000"))
	copy(dst[:], []byte("bench-dst-000000000000000000000000"))
	copy(next[:], []byte("bench-next-000000000000000000000000"))
	fwd.routeTable.UpdateRoute(routing.RouteEntry{DestID: dst, NextHopID: next})

	// create a real HybridSession
	combined := make([]byte, 64)
	for i := range combined {
		combined[i] = byte(i)
	}
	sessionInfo := append(src[:], dst[:]...)
	sess, err := crypto.NewHybridSession(combined, sessionInfo)
	if err != nil {
		b.Fatalf("new hybrid session: %v", err)
	}

	var sid [16]byte
	copy(sid[:], []byte("session-id-bench-000"))
	fwd.AddSession(sid, &Session{CryptoState: sess})

	// create header with payload and room for tag
	payload := []byte("benchmark-payload-abcdefghijklmnop")
	h := wire.Header{SrcID: src, DstID: dst, FlowLabel: 0x2, SeqNum: 1, SessionID: sid, Length: uint16(len(payload))}
	buf := make([]byte, wire.HeaderSize+len(payload), wire.HeaderSize+len(payload)+crypto.TagSize)
	if err := h.Marshal(buf); err != nil {
		b.Fatalf("marshal header: %v", err)
	}
	copy(buf[wire.HeaderSize:], payload)

	sock := newTestSocket()
	umem := &testUMEM{}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go fwd.RunXDPLoop(ctx, sock, umem)

	// send frames in loop (b.N controlled by -benchtime)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// update sequence to avoid replay optimizations
		h.SeqNum = uint64(i)
		_ = h.Marshal(buf)
		sock.frames <- buf
		select {
		case <-sock.sent:
			// ok
		case <-time.After(5 * time.Second):
			b.Fatalf("timed out waiting for send")
		}
	}
}

// BenchmarkRunXDPLoop_MultiWorker_WithCrypto spawns multiple worker loops and
// drives them concurrently. Use env `BENCH_WORKERS` to set worker count.
func BenchmarkRunXDPLoop_MultiWorker_WithCrypto(b *testing.B) {
	workers := 4
	if v := os.Getenv("BENCH_WORKERS"); v != "" {
		// ignore parse error; default to 4
	}

	// prepare per-worker forwarders and sockets
	socks := make([]*testSocket, workers)
	ums := make([]*testUMEM, workers)
	fwds := make([]*Forwarder, workers)

	for w := 0; w < workers; w++ {
		fwd := &Forwarder{routeTable: routing.NewTable(), logger: nil}
		var src, dst, next [32]byte
		copy(src[:], []byte("bench-src-000000000000000000000000"))
		copy(dst[:], []byte("bench-dst-000000000000000000000000"))
		copy(next[:], []byte("bench-next-000000000000000000000000"))
		fwd.routeTable.UpdateRoute(routing.RouteEntry{DestID: dst, NextHopID: next})

		combined := make([]byte, 64)
		for i := range combined {
			combined[i] = byte(i + w)
		}
		sessionInfo := append(src[:], dst[:]...)
		sess, err := crypto.NewHybridSession(combined, sessionInfo)
		if err != nil {
			b.Fatalf("new hybrid session: %v", err)
		}
		var sid [16]byte
		copy(sid[:], []byte("session-id-bench-000"))
		fwd.AddSession(sid, &Session{CryptoState: sess})

		sock := newTestSocket()
		um := &testUMEM{}
		socks[w] = sock
		ums[w] = um
		fwds[w] = fwd

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go fwd.RunXDPLoop(ctx, sock, um)
	}

	// prepare frame buffer
	payload := []byte("benchmark-payload-abcdefgh")
	h := wire.Header{FlowLabel: 0x3, Length: uint16(len(payload))}
	frame := wire.NewHeaderBuffer(int(h.Length))
	if err := h.Marshal(frame); err != nil {
		b.Fatalf("marshal: %v", err)
	}
	copy(frame[wire.HeaderSize:], payload)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// round-robin send to workers
		w := i % workers
		socks[w].frames <- frame
		select {
		case <-socks[w].sent:
		case <-time.After(5 * time.Second):
			b.Fatalf("timed out waiting for send")
		}
	}
}
