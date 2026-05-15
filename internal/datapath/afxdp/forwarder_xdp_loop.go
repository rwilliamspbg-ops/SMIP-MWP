//go:build withafxdp
// +build withafxdp

package afxdp

import (
	"context"
	"fmt"
)

// StartXDPForwarder initializes UMEM and socket resources and runs the core
// XDP loop. This function is only compiled when `withafxdp` is enabled and
// shows how production code will wire real resources to the testable loop.
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

	go f.RunXDPBatchLoop(ctx, sock, umem)
	return nil
}
