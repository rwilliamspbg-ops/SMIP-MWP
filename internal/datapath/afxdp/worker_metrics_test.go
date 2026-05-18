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
	"fmt"
	"sync"
	"testing"
	"time"

	"smip-mwp/internal/wire"

	"github.com/prometheus/client_golang/prometheus/testutil"
)

// TestPerWorkerMetrics spawns two short-lived workers using the test socket
// and asserts that per-worker metrics are incremented as frames are processed.
func TestPerWorkerMetrics(t *testing.T) {
	numWorkers := 2

	// Create test sockets and pre-load one frame per socket
	sockets := make([]*testSocket, numWorkers)
	for i := 0; i < numWorkers; i++ {
		sockets[i] = newTestSocket()
		// Build a minimal header buffer
		h := wire.Header{Length: 0}
		buf := wire.NewHeaderBuffer(0)
		if err := h.Marshal(buf); err != nil {
			t.Fatalf("marshal: %v", err)
		}
		sockets[i].frames <- buf
	}

	// Worker function will Poll the corresponding socket once and increment
	// per-worker metrics using IncRxWorker.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	SpawnPerCPUWorkers(ctx, numWorkers, &wg, func(wctx context.Context, id int) {
		// simple worker: poll socket once and increment metric
		frames, _ := sockets[id].Poll(1)
		if len(frames) > 0 {
			IncRxWorker(id, len(frames))
		}
	})

	// Wait for workers to finish (they'll exit after one run)
	wg.Wait()

	// Allow metric propagation
	time.Sleep(10 * time.Millisecond)

	// Assert metrics per worker
	for i := 0; i < numWorkers; i++ {
		label := fmt.Sprint(i)
		got := testutil.ToFloat64(rxPacketsVec.WithLabelValues(label))
		if int(got) != 1 {
			t.Fatalf("worker %d: expected rx 1, got %v", i, got)
		}
	}
}
