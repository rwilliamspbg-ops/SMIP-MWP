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

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"smip-mwp/internal/datapath/afxdp"
	"smip-mwp/internal/routing"
)

func main() {
	iface := flag.String("iface", "eth0", "network interface")
	flag.Parse()

	rt := routing.NewTable()
	var dest, next [32]byte
	copy(dest[:], []byte("example-dst-000000000000000000000"))
	copy(next[:], []byte("example-next-000000000000000000000"))
	rt.UpdateRoute(routing.RouteEntry{DestID: dest, NextHopID: next})
	cfg := afxdp.Config{
		Interface: *iface,
		QueueID:   0,
		ZeroCopy:  false,
		NumFrames: 2048,
		FrameSize: 2048,
		BatchSize: 64,
	}
	fwd, err := afxdp.NewForwarder(cfg, rt)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create forwarder: %v\n", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		cancel()
	}()

	go func() {
		fwd.Run(ctx)
	}()

	// Wait for cancellation
	<-ctx.Done()
	// Graceful shutdown with timeout
	done := make(chan struct{})
	go func() {
		fwd.Close()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		fmt.Println("shutdown timeout")
	}
}
