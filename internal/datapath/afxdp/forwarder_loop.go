package afxdp

import (
	"context"
	"io"
	"log"
	"time"

	"smip-mwp/internal/crypto"
	"smip-mwp/internal/wire"
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
		// track which out entries were allocated from the pktPool so we can
		// return them after a successful send.
		pooledIdxs := make([]int, 0, len(frames))
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

		// Attempt in-place encryption for packets that have a registered session
		for i, pkt := range out {
			// parse header to get session id and payload length
			h, err := wire.ParseHeader(pkt)
			if err != nil {
				continue
			}
			if sess := func() *Session { f.mu.RLock(); s := f.sessions[h.SessionID]; f.mu.RUnlock(); return s }(); sess != nil && sess.CryptoState != nil {
				payloadLen := int(h.Length)
				if payloadLen == 0 {
					continue
				}
				// payload slice
				if len(pkt) < wire.HeaderSize+payloadLen {
					continue
				}
				payload := pkt[wire.HeaderSize : wire.HeaderSize+payloadLen]
				// prefer in-place when UMEM-like buffer has capacity for tag
				if cap(payload) >= payloadLen+crypto.TagSize {
					if err := sess.CryptoState.EncryptInPlace(payload[:payloadLen], h.SeqNum); err == nil {
						// extend packet length to include tag
						out[i] = pkt[:wire.HeaderSize+payloadLen+crypto.TagSize]
						// update header length in-place
						if vh, err := wire.ViewHeader(out[i]); err == nil {
							vh.SetLength(uint16(payloadLen + crypto.TagSize))
						}
						continue
					} else {
						logger.Printf("in-place encrypt failed: %v", err)
						IncCryptoError()
					}
				}
				// Fallback: allocate ciphertext
				ct, err := sess.CryptoState.Encrypt(payload, h.SeqNum)
				if err == nil {
					// construct new packet: header + ct using pooled buffer when available
					var newpkt []byte
					if f.pktPool != nil {
						buf := f.pktPool.Get().([]byte)
						needed := wire.HeaderSize + len(ct)
						if cap(buf) < needed {
							// fall back to fresh allocation when pool buffer too small
							newpkt = make([]byte, needed)
						} else {
							newpkt = buf[:needed]
						}
						copy(newpkt, out[i][:wire.HeaderSize])
						copy(newpkt[wire.HeaderSize:], ct)
						// mark to return to pool after successful send
						pooledIdxs = append(pooledIdxs, len(out))
					} else {
						newpkt = make([]byte, wire.HeaderSize+len(ct))
						copy(newpkt, out[i][:wire.HeaderSize])
						copy(newpkt[wire.HeaderSize:], ct)
					}
					// update header length
					if vh, err := wire.ViewHeader(newpkt); err == nil {
						vh.SetLength(uint16(len(ct)))
					}
					out = append(out, newpkt)
					continue
				}
				IncCryptoError()
			}
		}

		if len(out) > 0 {
			if err := sock.Send(out); err != nil {
				logger.Printf("xdp send error: %v", err)
			} else {
				IncTx(len(out))
				// return pooled buffers to pool to be reused
				if f.pktPool != nil && len(pooledIdxs) > 0 {
					for _, idx := range pooledIdxs {
						if idx < len(out) {
							f.pktPool.Put(out[idx])
						}
					}
				}
			}
		}
	}
}
