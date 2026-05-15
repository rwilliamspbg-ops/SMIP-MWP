//go:build withafxdp
// +build withafxdp

package afxdp

import (
	"context"
	"encoding/binary"
	"fmt"
	"time"

	"smip-mwp/internal/crypto"
)

// RunXDPWorker implements a simple RX->process->TX worker using the real
// XDPSocket and UMEM. It decrypts incoming packets when a session exists and
// encrypts outgoing packets accordingly. This is a minimal example — real
// implementations must handle zero-copy buffers, MAC rewriting, and batching
// more carefully.
func (f *Forwarder) RunXDPWorker(ctx context.Context, sock *XDPSocket, umem *UMEM) {
	defer func() {
		if umem != nil {
			_ = umem.Close()
		}
		if sock != nil {
			_ = sock.Close()
		}
	}()

	poll := func() ([][]byte, error) {
		// Use the XDPSocket.Poll wrapper which adapts to the underlying
		// asavie/xdp socket APIs via reflection. This keeps the code resilient
		// across library versions.
		if sock == nil {
			return nil, fmt.Errorf("socket nil")
		}
		return sock.Poll(64)
	}

	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			f.logger.Println("xdp worker: stopping")
			return
		case <-ticker.C:
		}

		frames, err := poll()
		if err != nil {
			// In production, handle errors appropriately
			f.logger.Printf("xdp rx error: %v", err)
			time.Sleep(10 * time.Millisecond)
			continue
		}
		if len(frames) == 0 {
			continue
		}
		IncRx(len(frames))

		var out [][]byte
		for _, b := range frames {
			// Parse header; the header parsing helper expects a full buffer
			// where the header occupies the first bytes.
			hdr, err := PrepareForPacket(b, f.routeTable)
			_ = hdr // placeholder for future use
			if err != nil {
				f.logger.Printf("prepare packet failed: %v", err)
				IncDropped(1)
				continue
			}

			// Example: session lookup by the 16-byte SessionID at offset
			// We expect the session id to be embedded; parse it.
			if len(b) >= 94+2 { // ensure header.lenOffset exists
				// session id is 16 bytes at offset 76 (see header.go)
				var sid [16]byte
				copy(sid[:], b[76:92])

				f.mu.RLock()
				sess := f.sessions[sid]
				f.mu.RUnlock()

				if sess != nil && sess.CryptoState != nil {
					// If the buffer contains ciphertext (>= tag size), attempt decrypt.
					if len(b) >= crypto.TagSize {
						// DecryptInPlace expects a buffer that may include tag; make a copy
						// because UMEM-backed buffers may be special
						tmp := make([]byte, len(b))
						copy(tmp, b)
						pt, err := sess.CryptoState.DecryptInPlace(tmp, binary.BigEndian.Uint64(tmp[68:76]))
						if err == nil {
							out = append(out, pt)
							continue
						}
						// on error, fall through to pass-through
						f.logger.Printf("decrypt error: %v", err)
						IncCryptoError()
					}
				}
			}
			// Default: forward unchanged
			out = append(out, b)
		}
		// Prepare a transmit batch
		txBatch := make([][]byte, 0, len(out))
		for _, pkt := range out {
			// Inspect session id again
			if len(pkt) >= 92 {
				var sid [16]byte
				copy(sid[:], pkt[76:92])
				f.mu.RLock()
				sess := f.sessions[sid]
				f.mu.RUnlock()
				if sess != nil && sess.CryptoState != nil {
					// Encrypt (slow path) using AEAD
					ct, err := sess.CryptoState.Encrypt(pkt, uint64(time.Now().UnixNano()))
					if err == nil {
						txBatch = append(txBatch, ct)
					} else {
						f.logger.Printf("encrypt error: %v", err)
						IncCryptoError()
						// fall back to sending unencrypted packet
						txBatch = append(txBatch, pkt)
					}
					continue
				}
			}
			// send pkt as-is
			txBatch = append(txBatch, pkt)
		}

		if len(txBatch) > 0 {
			if err := sock.Send(txBatch); err != nil {
				f.logger.Printf("xdp tx error: %v", err)
			} else {
				IncTx(len(txBatch))
			}
		}
	}
}
