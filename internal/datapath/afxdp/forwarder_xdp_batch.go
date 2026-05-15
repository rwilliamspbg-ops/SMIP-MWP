//go:build withafxdp
// +build withafxdp

package afxdp

import (
	"context"
	"time"
)

// RunXDPBatchLoop runs a fully batched RX->process->TX loop reusing descriptor
// slices and UMEM frames to minimize allocations and maximize throughput.
func (f *Forwarder) RunXDPBatchLoop(ctx context.Context, sock *XDPSocket, umem *UMEM) {
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
		// avoid nil deref
		logger = nil
	}

	xsk := sock.s
	batchInterval := 50 * time.Millisecond
	ticker := time.NewTicker(batchInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			if logger != nil {
				logger.Println("xdp batch loop: stopping")
			}
			return
		case <-ticker.C:
		default:
		}

		// Refill UMEM fill ring with free descriptors
		free := xsk.NumFreeFillSlots()
		if free > 0 {
			descs := xsk.GetDescs(free)
			if len(descs) > 0 {
				xsk.Fill(descs)
			}
		}

		// Poll for RX/COMPLETION events
		numRx, numCompleted, err := xsk.Poll(100)
		if err != nil {
			if logger != nil {
				logger.Printf("xdp poll: %v", err)
			}
			time.Sleep(1 * time.Millisecond)
			continue
		}
		if numCompleted > 0 {
			xsk.Complete(numCompleted)
		}

		if numRx == 0 {
			continue
		}

		// Receive descriptors and process frames in-place
		descs := xsk.Receive(numRx)
		for i := 0; i < len(descs); i++ {
			d := descs[i]
			frame := xsk.GetFrame(d)

			// Process frame: steering and crypto handling (reuse existing helpers)
			// Note: PrepareForPacket expects a full header buffer
			if _, _, err := PrepareForPacket(frame, f.routeTable); err != nil {
				if logger != nil {
					logger.Printf("prepare for packet failed: %v", err)
				}
				IncDropped(1)
				continue
			}

			// Session lookup and in-place decrypt if needed (similar to earlier logic)
			if len(frame) >= HeaderSize+2 {
				// session id at offset sessionOffset (header.go)
				var sid [16]byte
				copy(sid[:], frame[sessionOffset:sessionOffset+16])
				f.mu.RLock()
				sess := f.sessions[sid]
				f.mu.RUnlock()
				if sess != nil && sess.CryptoState != nil {
					// header length
					hdr, err := ViewHeader(frame)
					if err == nil {
						payloadLen := int(hdr.Length())
						if payloadLen >= TagSize {
							// decrypt in-place on UMEM frame
							payload := frame[HeaderSize : HeaderSize+payloadLen]
							if _, err := sess.CryptoState.DecryptInPlace(payload, hdr.SeqNum()); err == nil {
								// update header length to plaintext length
								// DecryptInPlace returns plaintext slice length; we assume it shrinks by TagSize
								newLen := len(payload) - TagSize
								hdr.SetLength(uint16(newLen))
								d.Len = uint32(HeaderSize + newLen)
							} else {
								IncCryptoError()
								if logger != nil {
									logger.Printf("decrypt error: %v", err)
								}
							}
						}
					}
				}
			}
		}

		// Transmit all received descriptors (modified in-place)
		if len(descs) > 0 {
			xsk.Transmit(descs)
		}
	}
}
