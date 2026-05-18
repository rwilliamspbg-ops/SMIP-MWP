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
	"encoding/binary"
	"fmt"
	"log"
	"sync"

	"smip-mwp/internal/crypto"
	"smip-mwp/internal/routing"
)

// Config contains lightweight AF_XDP options used by the stub forwarder.
type Config struct {
	Interface string
	QueueID   int
	ZeroCopy  bool
	NumFrames int
	FrameSize int
	BatchSize int
	// Adaptive batch sizing bounds
	BatchSizeMin int
	BatchSizeMax int
	NumWorkers   int // number of per-CPU workers / queues to spawn (0 -> NumCPU)
	// FillThreshold controls how many descriptors we attempt to keep
	// available on the UMEM Fill ring. If zero, defaults to BatchSize.
	FillThreshold int
	// Adaptive fill controls
	// If true, dynamically adjust FillThreshold based on observed completion rate.
	FillAdaptive bool
	// Multiplicative factor applied to the observed completion rate to compute target fill.
	FillAdaptFactor float64
	// EMA alpha used to smooth observed completion rate (0..1). Higher alpha weights recent samples.
	FillEMAAlpha float64
	// Minimum and maximum allowed fill target when adaptive mode is enabled.
	FillMin int
	FillMax int
}

// Session represents a lightweight session placeholder.
type Session struct {
	CryptoState *crypto.HybridSession
	FlowLabel   uint32
}

// Forwarder is a minimal stub implementation that satisfies the public API used by cmd.
type Forwarder struct {
	cfg        Config
	logger     *log.Logger
	routeTable *routing.Table
	running    bool

	// Sharded session map to reduce global RWMutex contention. Use a fixed
	// number of shards (power-of-two recommended) and hash the first 8
	// bytes of the session ID to select a shard.
	sessionShards [16]struct {
		sessions map[[16]byte]*Session
		mu       sync.RWMutex
	}
	// pktPool supplies buffers for constructing fallback packets to reduce
	// per-worker pkt pools (initialized in Start) to avoid global sync.Pool contention.
	// If nil, legacy `pktPool` may be used.
	workerPools []*workerPktPool
	// pktPool supplies buffers for constructing fallback packets to reduce
	// per-packet allocations in the hot path. Buffers are sized to `cfg.FrameSize`.
	pktPool *sync.Pool
	// worker lifecycle
	workersWG     sync.WaitGroup
	workersCancel context.CancelFunc
}

const numSessionShards = 16

// initSessionShards ensures internal maps are allocated for each shard.
func (f *Forwarder) initSessionShards() {
	for i := 0; i < numSessionShards; i++ {
		if f.sessionShards[i].sessions == nil {
			f.sessionShards[i].sessions = make(map[[16]byte]*Session)
		}
	}
}

func (f *Forwarder) getShardIndex(sid [16]byte) int {
	return int(binary.BigEndian.Uint64(sid[:8]) % uint64(numSessionShards))
}

// GetSession returns a session pointer or nil if not present. This is the
// preferred hot-path accessor to avoid global locks.
func (f *Forwarder) GetSession(sid [16]byte) *Session {
	idx := f.getShardIndex(sid)
	shard := &f.sessionShards[idx]
	shard.mu.RLock()
	s := shard.sessions[sid]
	shard.mu.RUnlock()
	return s
}

// Run executes the forwarder loop until context cancellation.
// For AF_XDP mode (withafxdp), spawns multi-queue workers.
// For stub mode, runs a simple polling loop.
func (f *Forwarder) Run(ctx context.Context) {
	f.running = true
	defer func() {
		f.running = false
		f.logger.Println("forwarder stopped")
	}()

	// In AF_XDP mode, Start() will spawn workers
	// In stub mode, this runs locally
	f.Start(ctx)

	// Wait for context cancellation
	<-ctx.Done()
}

// Close shuts down the forwarder.
func (f *Forwarder) Close() error {
	if f.running {
		// Best-effort stop
		f.running = false
		f.logger.Println("closed")
	}
	// stop workers if running
	f.Stop()
	return nil
}

// Stop cancels started workers and waits for them to exit.
func (f *Forwarder) Stop() {
	if f.workersCancel != nil {
		f.workersCancel()
	}
	f.workersWG.Wait()
}

// GetStats returns a small stats stub.
func (f *Forwarder) GetStats() (rx, tx, dropped uint64) {
	return 0, 0, 0
}

// Helper for demonstration
func (f *Forwarder) String() string { return fmt.Sprintf("afxdp.Forwarder(iface=%s)", f.cfg.Interface) }

// AddSession registers a session for a given session ID and records a handshake metric.
func (f *Forwarder) AddSession(sid [16]byte, s *Session) {
	idx := f.getShardIndex(sid)
	shard := &f.sessionShards[idx]
	shard.mu.Lock()
	if shard.sessions == nil {
		shard.sessions = make(map[[16]byte]*Session)
	}
	shard.sessions[sid] = s
	shard.mu.Unlock()
	IncHandshake()
}

// RemoveSession removes a session by its session ID.
func (f *Forwarder) RemoveSession(sid [16]byte) {
	idx := f.getShardIndex(sid)
	shard := &f.sessionShards[idx]
	shard.mu.Lock()
	delete(shard.sessions, sid)
	shard.mu.Unlock()
}
