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
	"testing"
	"time"

	"smip-mwp/internal/routing"
	"smip-mwp/internal/wire"
)

// BenchmarkRunXDPLoop_NoCrypto measures the lightweight receive->forward loop
// without performing in-path crypto. It uses the in-repo test socket/umem
// doubles to provide a repeatable benchmark in CI/dev environments.
func BenchmarkRunXDPLoop_NoCrypto(b *testing.B) {
	fwd := &Forwarder{routeTable: routing.NewTable(), logger: nil}

	var src, dst, next [32]byte
	copy(src[:], []byte("bench-src-000000000000000000000000"))
	copy(dst[:], []byte("bench-dst-000000000000000000000000"))
	copy(next[:], []byte("bench-next-000000000000000000000000"))
	fwd.routeTable.UpdateRoute(routing.RouteEntry{DestID: dst, NextHopID: next})

	h := wire.Header{SrcID: src, DstID: dst, FlowLabel: 0x1, Length: 0}
	buf := wire.NewHeaderBuffer(int(h.Length))
	if err := h.Marshal(buf); err != nil {
		b.Fatalf("marshal header: %v", err)
	}

	sock := newTestSocket()
	umem := &testUMEM{}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go fwd.RunXDPLoop(ctx, sock, umem)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sock.frames <- buf
		select {
		case <-sock.sent:
			// ok
		case <-time.After(2 * time.Second):
			b.Fatalf("timed out waiting for send")
		}
	}
}
