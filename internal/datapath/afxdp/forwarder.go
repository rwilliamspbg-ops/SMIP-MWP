package afxdp

import (
	"context"
	"fmt"
	"log"
	"sync"

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

	// sessions holds per-session crypto state keyed by SessionID (16 bytes).
	sessions map[[16]byte]*Session
	mu       sync.RWMutex
	// pktPool supplies buffers for constructing fallback packets to reduce
	// per-packet allocations in the hot path. Buffers are sized to `cfg.FrameSize`.
	pktPool *sync.Pool
	// worker lifecycle
	workersWG     sync.WaitGroup
	workersCancel context.CancelFunc
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
