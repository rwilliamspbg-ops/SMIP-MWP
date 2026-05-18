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
)

// StartXDPForwarder initializes and runs AF_XDP single-worker forwarder.
// For multi-worker setup, use Start() which spawns workers via SpawnPerCPUWorkers.
func (f *Forwarder) StartXDPForwarder(ctx context.Context) error {
	umem, err := NewUMEM(f.cfg.NumFrames, f.cfg.FrameSize)
	if err != nil {
		return fmt.Errorf("NewUMEM: %w", err)
	}

	sock, err := NewXDPSocket(f.cfg.Interface, f.cfg.QueueID, umem)
	if err != nil {
		umem.Close()
		return fmt.Errorf("NewXDPSocket: %w", err)
	}

	// Run single worker (workerID=0) in background
	workerID := 0
	go f.RunXDPBatchLoop(ctx, sock, umem, workerID)
	return nil
}
