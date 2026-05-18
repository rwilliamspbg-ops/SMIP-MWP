//go:build withafxdp
// +build withafxdp

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
