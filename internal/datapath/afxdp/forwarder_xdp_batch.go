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

	// EMA of completed descriptors per tick used when adaptive fill enabled.
	var emaCompleted float64
	alpha := 0.25 // default alpha; may be overridden by config below

	// Apply config defaults if not set
	if f.cfg.FillEMAAlpha > 0 {
		alpha = f.cfg.FillEMAAlpha
	}

	// Pre-allocate descriptor buffer to avoid per-poll allocations.
	descBuf := make([]*XDPDescriptor, 0, batchSize)

	// worker-local session cache
	var wcache workerSessionCache

	for {
		select {
		case <-ctx.Done():
			logger.Println("batch loop: stopping")
			return
		case <-ticker.C:
		default:
		}

		// Refill UMEM fill ring with free descriptors (minimize stalls).
		// If adaptive fill is enabled, compute desired based on EMA of completed descriptors.
		free := xsk.NumFreeFillSlots()
		if free > 0 {
			var desired int
			if f.cfg.FillAdaptive {
				// target = max(batchSize, int(ema*factor))
				factor := f.cfg.FillAdaptFactor
				if factor <= 0 {
					factor = 1.5
				}
				target := int(emaCompleted * factor)
				if target < batchSize {
					target = batchSize
				}
				// clamp to configured min/max when provided
				if f.cfg.FillMin > 0 && target < f.cfg.FillMin {
					target = f.cfg.FillMin
				}
				if f.cfg.FillMax > 0 && target > f.cfg.FillMax {
					target = f.cfg.FillMax
				}
				desired = target
				// export metrics for visibility
				SetFillTarget(desired)
			} else {
				desired = f.cfg.FillThreshold
				if desired <= 0 {
					desired = batchSize
				}
				SetFillTarget(desired)
			}
			if desired > free {
				desired = free
			}
			descs := xsk.GetDescs(desired)
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

		// Update EMA with recent completions for adaptive fill control
		if f.cfg.FillAdaptive {
			emaCompleted = alpha*float64(numCompleted) + (1.0-alpha)*emaCompleted
			SetFillEMA(emaCompleted)
		}

		if numRx == 0 {
			continue
		}

		// Receive descriptors and reuse pre-allocated buffer to avoid allocations
		descBuf = descBuf[:0]
		tmp := xsk.Receive(numRx)
		if len(tmp) == 0 {
			continue
		}
		descBuf = append(descBuf, tmp...)
		descs := descBuf
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

			// fast path: check worker-local cache
			sess := wcache.Get(sid)
			if sess == nil {
				sess = f.GetSession(sid)
				if sess != nil {
					wcache.Put(sid, sess)
				}
			}

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
