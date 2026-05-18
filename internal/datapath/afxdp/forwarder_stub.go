//go:build !withafxdp
// +build !withafxdp

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
	f := &Forwarder{cfg: cfg, logger: l, routeTable: routeTable, sessions: make(map[[16]byte]*Session)}
	// initialize pktPool sized to FrameSize (fallback to 2048 if unset)
	size := cfg.FrameSize
	if size <= 0 {
		size = 2048
	}
	f.pktPool = &sync.Pool{New: func() interface{} { return make([]byte, size) }}
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
