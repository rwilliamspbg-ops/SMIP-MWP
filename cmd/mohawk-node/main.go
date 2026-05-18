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

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"smip-mwp/internal/datapath/afxdp"
	"smip-mwp/internal/routing"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	iface := flag.String("iface", "eth0", "network interface")
	dry := flag.Bool("dry-run", true, "don't initialize AF_XDP; show plan only")
	metricsAddr := flag.String("metrics-addr", ":9090", "address to expose Prometheus metrics (empty to disable)")
	numFrames := flag.Int("frames", 4096, "number of UMEM frames (AF_XDP only)")
	frameSize := flag.Int("frame-size", 2048, "frame size in bytes (AF_XDP only)")
	batchSize := flag.Int("batch-size", 64, "packet batch size (AF_XDP only)")
	numWorkers := flag.Int("workers", 0, "number of worker threads (0 = num CPU cores)")
	zeroCopy := flag.Bool("zero-copy", true, "enable zero-copy mode for AF_XDP")
	flag.Parse()

	// Create routing table and prime with a sample route
	rt := routing.NewTable()
	var dest, next [32]byte
	copy(dest[:], []byte("example-dst-000000000000000000000"))
	copy(next[:], []byte("example-next-000000000000000000000"))
	rt.UpdateRoute(routing.RouteEntry{DestID: dest, NextHopID: next})

	fmt.Println("Routing table primed with sample entry")

	// Start metrics server if configured
	if *metricsAddr != "" {
		go func() {
			http.Handle("/metrics", promhttp.Handler())
			log.Printf("metrics: listening on %s\n", *metricsAddr)
			if err := http.ListenAndServe(*metricsAddr, nil); err != nil {
				log.Printf("metrics endpoint failed: %v", err)
			}
		}()
	}

	if *dry {
		fmt.Printf("Dry run: would create AF_XDP forwarder on %s\n", *iface)
		fmt.Printf("  Interface:    %s\n", *iface)
		fmt.Printf("  Frames:       %d\n", *numFrames)
		fmt.Printf("  Frame size:   %d bytes\n", *frameSize)
		fmt.Printf("  Batch size:   %d packets\n", *batchSize)
		fmt.Printf("  Workers:      %d\n", *numWorkers)
		fmt.Printf("  Zero-copy:    %v\n", *zeroCopy)
		fmt.Println("\nTo run with AF_XDP, rebuild with -tags=withafxdp and ensure kernel libs are installed:")
		fmt.Println("  go run -tags=withafxdp ./cmd/mohawk-node --iface=eth0 --dry-run=false")
		return
	}

	// Create AF_XDP forwarder configuration
	cfg := afxdp.Config{
		Interface:  *iface,
		QueueID:    0, // Start with queue 0
		ZeroCopy:   *zeroCopy,
		NumFrames:  *numFrames,
		FrameSize:  *frameSize,
		BatchSize:  *batchSize,
		NumWorkers: *numWorkers,
	}

	// Create forwarder
	fwd, err := afxdp.NewForwarder(cfg, rt)
	if err != nil {
		log.Fatalf("failed to create forwarder: %v", err)
	}

	// Set up graceful shutdown on SIGINT/SIGTERM
	ctx, cancel := context.WithCancel(context.Background())
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start forwarder in background
	go func() {
		<-sigChan
		log.Println("Shutdown signal received, stopping forwarder...")
		cancel()
	}()

	log.Printf("Starting AF_XDP forwarder on %s\n", *iface)
	log.Printf("Configuration: %d frames, %d byte frames, %d packet batch, %d workers",
		*numFrames, *frameSize, *batchSize, *numWorkers)

	// Run forwarder with 30-second startup timeout
	startCtx, startCancel := context.WithTimeout(ctx, 30*time.Second)

	go fwd.Run(startCtx)
	startCancel()

	// Wait for context cancellation (from signal handler)
	<-ctx.Done()

	// Clean up
	if err := fwd.Close(); err != nil {
		log.Printf("error closing forwarder: %v", err)
	}

	log.Println("Forwarder shutdown complete")
}
