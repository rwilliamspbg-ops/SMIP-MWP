package afxdp

import (
	"context"
	"io"
	"log"
	"time"
)

// Abstractions used by the XDP loop so tests can provide test doubles.
type xdpUMEM interface {
	Close() error
}

type xdpSocket interface {
	// Poll returns up to `max` received frames as byte slices. If no frames are
	// available, it may block briefly or return an empty slice.
	Poll(max int) ([][]byte, error)
	// Send transmits a batch of packets. For test doubles this can record
	// packets for assertion.
	Send(pkts [][]byte) error
	Close() error
}

// RunXDPLoop runs the receive->steer->transmit loop using provided socket and
// umem abstractions. It is intentionally conservative and safe for CI/dev
// environments.
func (f *Forwarder) RunXDPLoop(ctx context.Context, sock xdpSocket, umem xdpUMEM) {
	defer func() {
		if umem != nil {
			_ = umem.Close()
		}
		if sock != nil {
			_ = sock.Close()
		}
	}()

	// Ensure non-nil logger for tests
	logger := f.logger
	if logger == nil {
		logger = log.New(io.Discard, "", 0)
	}

	pollBatch := 64
	tick := time.NewTicker(100 * time.Millisecond)
	defer tick.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Println("xdp loop: context cancelled")
			return
		case <-tick.C:
			// periodic wake to check for frames
		default:
		}

		frames, err := sock.Poll(pollBatch)
		if err != nil {
			// On transient errors, back off briefly and continue.
			logger.Printf("xdp poll error: %v", err)
			time.Sleep(10 * time.Millisecond)
			continue
		}
		if len(frames) == 0 {
			// No work; small sleep to avoid busy loop
			time.Sleep(5 * time.Millisecond)
			continue
		}

		// record rx
		IncRx(len(frames))

		out := make([][]byte, 0, len(frames))
		for _, buf := range frames {
			// Parse and select next-hop/queue
			nh, q, err := PrepareForPacket(buf, f.routeTable)
			if err != nil {
				logger.Printf("prepare for packet failed: %v", err)
				IncDropped(1)
				continue
			}
			_ = nh
			_ = q
			// For now, forward the original buffer (in real code we'll rewrite MACs
			// and potentially encrypt). Tests only assert that Send is called.
			out = append(out, buf)
		}

		if len(out) > 0 {
			if err := sock.Send(out); err != nil {
				logger.Printf("xdp send error: %v", err)
			} else {
				IncTx(len(out))
			}
		}
	}
}
