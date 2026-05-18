// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright 2026 rwilliamspbg-ops
//
// This file is part of SMIP-MWP.
// SMIP-MWP is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation; either version 3 of the License, or (at your option) any later version.
// See the LICENSE file in the project root for details.

package afxdp

import "testing"

func TestWorkerPktPoolGetNilReceiver(t *testing.T) {
	var pool *workerPktPool
	buf := pool.Get()
	if buf == nil {
		t.Fatal("expected nil receiver to return a non-nil buffer slice")
	}
	if len(*buf) != 0 {
		t.Fatalf("expected empty buffer, got len=%d", len(*buf))
	}
}

func TestWorkerPktPoolPutResetsLength(t *testing.T) {
	pool := newWorkerPktPool(128, 0)
	buf := pool.Get()
	if buf == nil {
		t.Fatal("expected buffer")
	}
	*buf = append(*buf, 1, 2, 3, 4)
	pool.Put(buf)

	reused := pool.Get()
	if reused == nil {
		t.Fatal("expected reused buffer")
	}
	if len(*reused) != 0 {
		t.Fatalf("expected zero-length buffer after Put/Get cycle, got len=%d", len(*reused))
	}
	if cap(*reused) < 128 {
		t.Fatalf("expected capacity >= 128, got %d", cap(*reused))
	}
}
