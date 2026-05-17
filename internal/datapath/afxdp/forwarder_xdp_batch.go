//go:build withafxdp
// +build withafxdp

package afxdp

import (
	"context"
	"log"
	"time"

	"smip-mwp/internal/crypto"
	"smip-mwp/internal/wire"
)

// RunXDPBatchLoop runs a fully batched RX->process->TX loop reusing descriptor
// slices and UMEM frames to minimize allocations and maximize throughput.
// This is the high-performance datapath optimized for 10 Gbps.
func (f *Forwarder) RunXDPBatchLoop(ctx context.Context, sock *XDPSocket, umem *UMEM, workerID int) {
	defer func() {
		if umem != nil {
			_ = umem.Close()
		}
		if sock != nil {
			_ = sock.Close()
		}
	}()

	logger := f.logger
	if logger == nil {
		logger = log.New(log.Writer(), "afxdp:batch:", 0)
	}

	// Get underlying xdp socket for batch operations
	xsk := sock.s
	if xsk == nil {
		logger.Println("socket not initialized")
		return
	}

	// Batch processing parameters (tuned for 10 Gbps)
	batchSize := f.cfg.BatchSize
	if batchSize <= 0 {
		batchSize = 64
	}
	batchInterval := 10 * time.Millisecond // Reduced for lower latency
	ticker := time.NewTicker(batchInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Println("batch loop: stopping")
			return
		case <-ticker.C:
		default:
		}

		// Refill UMEM fill ring with free descriptors (minimize stalls)
		free := xsk.NumFreeFillSlots()
		if free > 0 {
			descs := xsk.GetDescs(free)
			if len(descs) > 0 {
				xsk.Fill(descs)
			}
		}

		// Poll for RX/COMPLETION events (non-blocking)
		numRx, numCompleted, err := xsk.Poll(batchSize)
		if err != nil {
			logger.Printf("poll error: %v", err)
			time.Sleep(1 * time.Millisecond)
			continue
		}

		// Complete transmitted descriptors (return to pool)
		if numCompleted > 0 {
			xsk.Complete(numCompleted)
			IncTxWorker(workerID, numCompleted)
		}

		if numRx == 0 {
			continue
		}

		// Receive descriptors and process frames in-place (zero-copy hot path)
		descs := xsk.Receive(numRx)
		IncRxWorker(workerID, len(descs))
		start := time.Now()

		// Batch-process all received descriptors
		for i := 0; i < len(descs); i++ {
			d := descs[i]
			frame := xsk.GetFrame(d)

			// Minimal frame validation
			if len(frame) < wire.HeaderSize {
				IncDroppedWorker(workerID, 1)
				continue
			}

			// Parse header to get session ID (zero-copy view)
			hdr, err := wire.ViewHeader(frame)
			if err != nil {
				logger.Printf("header parse error: %v", err)
				IncDroppedWorker(workerID, 1)
				continue
			}

			// Session lookup and in-place decryption if needed
			sessionID := hdr.SessionID()
			var sid [16]byte
			copy(sid[:], sessionID)

			f.mu.RLock()
			sess := f.sessions[sid]
			f.mu.RUnlock()

			if sess != nil && sess.CryptoState != nil {
				payloadLen := int(hdr.Length())
				if payloadLen >= crypto.TagSize {
					// In-place decryption on UMEM frame (zero-copy hot path)
					payload := frame[wire.HeaderSize : wire.HeaderSize+payloadLen]
					if _, err := sess.CryptoState.DecryptInPlace(payload, hdr.SeqNum()); err == nil {
						// Update header length to plaintext length (shrunk by tag size)
						newLen := len(payload) - crypto.TagSize
						hdr.SetLength(uint16(newLen))
						d.Len = uint32(wire.HeaderSize + newLen)
					} else {
						IncCryptoError()
						logger.Printf("decrypt error: %v", err)
						IncDroppedWorker(workerID, 1)
						continue
					}
				}
			}
		}

		// Transmit all processed descriptors (modified in-place, zero-copy)
		if len(descs) > 0 {
			xsk.Transmit(descs)
			ObserveProcessingLatency(workerID, time.Since(start).Seconds())
		}
	}
}
