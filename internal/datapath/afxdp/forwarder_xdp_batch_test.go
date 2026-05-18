//go:build withafxdp
// +build withafxdp

// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright 2026 rwilliamspbg-ops
//
// This file is part of SMIP-MWP.
// SMIP-MWP is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or (at your
// option) any later version.
// See the LICENSE file in the project root for details.

package afxdp

import (
	"context"
	"testing"
	"time"

	"smip-mwp/internal/wire"
)

func TestRunXDPBatchLoop_TruncatedEncryptedFrameDoesNotPanic(t *testing.T) {
	umem, err := NewUMEM(1, wire.HeaderSize)
	if err != nil {
		t.Fatalf("new umem: %v", err)
	}
	frame := umem.GetFrameByIndex(0)
	if len(frame) != wire.HeaderSize {
		t.Fatalf("unexpected frame size: %d", len(frame))
	}

	var src, dst, next [32]byte
	copy(src[:], []byte("batch-src-000000000000000000000000"))
	copy(dst[:], []byte("batch-dst-000000000000000000000000"))
	copy(next[:], []byte("batch-next-00000000000000000000000"))

	h := wire.Header{SrcID: src, DstID: dst, SeqNum: 1, Length: uint16(wire.HeaderSize + 32)}
	if err := h.Marshal(frame); err != nil {
		t.Fatalf("marshal header: %v", err)
	}

	xsk := &xdpSocketImpl{umem: umem, rxRing: []*XDPDescriptor{{Addr: 0, Len: uint32(len(frame))}}}
	sock := &XDPSocket{s: xsk}
	fwd := &Forwarder{}

	ctx, cancel := context.WithCancel(context.Background())
	go fwd.RunXDPBatchLoop(ctx, sock, umem, 0)

	time.Sleep(50 * time.Millisecond)
	cancel()
	time.Sleep(50 * time.Millisecond)
}
