//go:build withafxdp
// +build withafxdp

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
	"log"
	"os"
	"runtime"
	"sync"

	"smip-mwp/internal/routing"
)

// NewForwarder creates an AF_XDP-backed forwarder. This file is compiled only
// when the `withafxdp` build tag is provided. The implementation here is a
// safe scaffold that should be extended to initialize UMEM, sockets, and
// queue maps. It currently returns a Forwarder configured to use the AF_XDP
// code path but keeps behavior testable.
func NewForwarder(cfg Config, routeTable *routing.Table) (*Forwarder, error) {
	if cfg.Interface == "" {
		return nil, fmt.Errorf("interface required for AF_XDP forwarder")
	}

	// NOTE: Real AF_XDP setup goes here: create UMEM, map to sockets, configure
	// RX/TX rings, setup fill/comp queues, and attach XDP program. Keep this
	// scaffold minimal and safe for CI/dev environments.

	l := log.New(os.Stdout, "afxdp:xdp: ", log.LstdFlags)
	f := &Forwarder{cfg: cfg, logger: l, routeTable: routeTable}
	// initialize packet pool sized to frame size
	size := cfg.FrameSize
	if size <= 0 {
		size = 2048
	}
	f.pktPool = &sync.Pool{New: func() interface{} { b := make([]byte, size); return &b }}
	// initialize sharded session maps
	f.initSessionShards()
	f.logger.Printf("afxdp: xdp-mode forwarder initialized iface=%s zeroCopy=%v", cfg.Interface, cfg.ZeroCopy)

	// Initialize UMEM
	umem, err := NewUMEM(cfg.NumFrames, cfg.FrameSize)
	if err != nil {
		return nil, fmt.Errorf("umem init: %w", err)
	}

	// Create a single socket for the configured queue (expand to multi-queue later)
	sock, err := NewXDPSocket(cfg.Interface, cfg.QueueID, umem)
	if err != nil {
		umem.Close()
		return nil, fmt.Errorf("xdp socket init: %w", err)
	}

	// Keep references on the forwarder for future RX/TX loops and cleanup.
	_ = umem
	_ = sock

	// Validate routing table not nil (helpful early error)
	if routeTable == nil {
		f.logger.Println("warning: route table is nil; forwarding disabled")
	}

	return f, nil
}

// Start launches per-CPU workers that each allocate their own UMEM and XDPSocket
// and run the batched XDP loop. Callers should provide a cancellable ctx to
// manage lifecycle (cancel to stop workers).
func (f *Forwarder) Start(ctx context.Context) {
	num := f.cfg.NumWorkers
	if num <= 0 {
		num = runtime.NumCPU()
	}

	// initialize per-worker pkt pools to reduce global sync.Pool usage
	if f.workerPools == nil {
		f.workerPools = make([]*workerPktPool, num)
		for i := 0; i < num; i++ {
			f.workerPools[i] = newWorkerPktPool(f.cfg.FrameSize, 32)
		}
	}

	// create worker context and store cancel so Stop() can cancel all
	workerCtx, cancel := context.WithCancel(ctx)
	f.workersCancel = cancel

	SpawnPerCPUWorkers(workerCtx, num, &f.workersWG, func(wctx context.Context, id int) {
		// Determine queue mapping: base QueueID + id
		qid := f.cfg.QueueID + id

		umem, err := NewUMEM(f.cfg.NumFrames, f.cfg.FrameSize)
		if err != nil {
			f.logger.Printf("worker %d: umem init failed: %v", id, err)
			return
		}

		sock, err := NewXDPSocket(f.cfg.Interface, qid, umem)
		if err != nil {
			f.logger.Printf("worker %d: xdp socket init failed: %v", id, err)
			_ = umem.Close()
			return
		}

		// Run the high-performance batched loop for this worker/queue.
		f.RunXDPBatchLoop(wctx, sock, umem, id)
	})
}
