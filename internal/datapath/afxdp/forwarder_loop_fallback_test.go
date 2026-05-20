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
	stdRand "crypto/rand"
	"testing"
	"time"

	"smip-mwp/internal/crypto"
	"smip-mwp/internal/routing"
	"smip-mwp/internal/wire"
)

func TestRunXDPLoop_FallbackEncryptReplacesPacket(t *testing.T) {
	fwd := &Forwarder{routeTable: routing.NewTable()}

	var src, dst [32]byte
	copy(src[:], []byte("fallback-src-0000000000000000000000"))
	copy(dst[:], []byte("fallback-dst-0000000000000000000000"))

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
	copy(sid[:], []byte("session-id-fallback"))
	fwd.AddSession(sid, &Session{CryptoState: sess})

	var next [32]byte
	copy(next[:], []byte("next-hop-fallback-0000000000000000"))
	fwd.routeTable.UpdateRoute(routing.RouteEntry{DestID: dst, NextHopID: next})

	payload := []byte("fallback encryption path")
	h := wire.Header{SrcID: src, DstID: dst, FlowLabel: 0x3, SeqNum: 7, SessionID: sid, Length: uint16(len(payload))}
	buf := make([]byte, wire.HeaderSize+len(payload))
	if err := h.Marshal(buf); err != nil {
		t.Fatalf("marshal header: %v", err)
	}
	copy(buf[wire.HeaderSize:], payload)

	sock := newTestSocket()
	umem := &testUMEM{}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go fwd.RunXDPLoop(ctx, sock, umem)
	defer cancel()

	sock.frames <- buf

	select {
	case <-sock.sentSignal:
		sock.mu.Lock()
		sent := sock.lastSent
		sock.mu.Unlock()
		if len(sent) != 1 {
			t.Fatalf("expected exactly 1 packet, got %d", len(sent))
		}
		pkt := sent[0]
		ph, err := wire.ParseHeader(pkt)
		if err != nil {
			t.Fatalf("parse sent header: %v", err)
		}
		if int(ph.Length) != len(payload)+crypto.TagSize {
			t.Fatalf("expected payload length %d, got %d", len(payload)+crypto.TagSize, ph.Length)
		}
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
