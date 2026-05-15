//go:build withafxdp
// +build withafxdp

package afxdp

import (
	"fmt"
	"log"
	"os"

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
	f := &Forwarder{cfg: cfg, logger: l, routeTable: routeTable, sessions: make(map[[16]byte]*Session)}
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
