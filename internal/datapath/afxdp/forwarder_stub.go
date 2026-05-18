//go:build !withafxdp
// +build !withafxdp

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
	"log"
	"os"
	"sync"
	"time"

	"smip-mwp/internal/routing"
)

// NewForwarder creates the non-AF_XDP stub forwarder used for unit tests and
// development when the withafxdp build tag is not set.
func NewForwarder(cfg Config, routeTable *routing.Table) (*Forwarder, error) {
	l := log.New(os.Stdout, "afxdp: ", log.LstdFlags)
	f := &Forwarder{cfg: cfg, logger: l, routeTable: routeTable}
	// initialize pktPool sized to FrameSize (fallback to 2048 if unset)
	size := cfg.FrameSize
	if size <= 0 {
		size = 2048
	}
	f.pktPool = &sync.Pool{New: func() interface{} { b := make([]byte, size); return &b }}
	// initialize sharded session maps
	f.initSessionShards()
	f.logger.Printf("stub forwarder created iface=%s", cfg.Interface)
	return f, nil
}

// Start is a stub implementation for non-AF_XDP mode
func (f *Forwarder) Start(ctx context.Context) {
	f.logger.Println("stub mode: Start() called but no workers spawned")
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			f.logger.Println("stub mode: context cancelled")
			return
		case <-ticker.C:
			// Periodic status update
			f.logger.Printf("stub tick: %v", f.running)
		}
	}
}
