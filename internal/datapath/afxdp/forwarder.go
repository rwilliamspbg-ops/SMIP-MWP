package afxdp

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"smip-mwp/internal/crypto"
	"smip-mwp/internal/routing"
)

// Config contains lightweight AF_XDP options used by the stub forwarder.
type Config struct {
	Interface  string
	QueueID    int
	ZeroCopy   bool
	NumFrames  int
	FrameSize  int
	BatchSize  int
	NumWorkers int // number of per-CPU workers / queues to spawn (0 -> NumCPU)
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

	// sessions holds per-session crypto state keyed by SessionID (16 bytes).
	sessions map[[16]byte]*Session
	mu       sync.RWMutex
}

// Run executes the forwarder loop (stub) until context cancellation.
func (f *Forwarder) Run(ctx context.Context) {
	f.running = true
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			f.logger.Println("forwarder stopping")
			f.running = false
			return
		case <-ticker.C:
			// Periodic noop to show status in logs
			f.logger.Printf("tick running=%v", f.running)
		}
	}
}

// Close shuts down the forwarder.
func (f *Forwarder) Close() error {
	if f.running {
		// Best-effort stop
		f.running = false
		f.logger.Println("closed")
	}
	return nil
}

// GetStats returns a small stats stub.
func (f *Forwarder) GetStats() (rx, tx, dropped uint64) {
	return 0, 0, 0
}

// Helper for demonstration
func (f *Forwarder) String() string { return fmt.Sprintf("afxdp.Forwarder(iface=%s)", f.cfg.Interface) }

// AddSession registers a session for a given session ID and records a handshake metric.
func (f *Forwarder) AddSession(sid [16]byte, s *Session) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.sessions == nil {
		f.sessions = make(map[[16]byte]*Session)
	}
	f.sessions[sid] = s
	IncHandshake()
}

// RemoveSession removes a session by its session ID.
func (f *Forwarder) RemoveSession(sid [16]byte) {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.sessions, sid)
}
