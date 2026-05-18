// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright 2026 rwilliamspbg-ops
//
// This file is part of SMIP-MWP.
// SMIP-MWP is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation; either version 3 of the License, or (at your option) any later version.
// See the LICENSE file in the project root for details.

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

	// per-goroutine worker-local session cache to avoid hot map lookups
	var wcache workerSessionCache
	// per-goroutine packet pool to avoid sync.Pool on hot path
	wpool := newWorkerPktPool(f.cfg.FrameSize, 0)

	for {
		// Fast check for cancellation without using `select` inside the hot loop.
		if ctx.Err() != nil {
			logger.Println("xdp loop: context cancelled")
			return
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
		// return them after a successful send. We store pointers to pooled
		// buffers (type *[]byte) to avoid passing pointer-like values by value.
		pooledPtrs := make([]*[]byte, 0, len(frames))
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
			// perform session lookup via worker-local cache then sharded map
			var sid [16]byte
			copy(sid[:], h.SessionID[:])
			sess := wcache.Get(sid)
			if sess == nil {
				sess = f.GetSession(sid)
				if sess != nil {
					wcache.Put(sid, sess)
				}
			}
			if sess != nil && sess.CryptoState != nil {
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
					if wpool != nil {
						bufPtr := wpool.Get()
						needed := wire.HeaderSize + len(ct)
						if cap(*bufPtr) < needed {
							// fall back to fresh allocation when pool buffer too small
							wpool.Put(bufPtr)
							newpkt = make([]byte, needed)
						} else {
							newpkt = (*bufPtr)[:needed]
							// mark this pool pointer for later return
							pooledPtrs = append(pooledPtrs, bufPtr)
						}
						copy(newpkt, out[i][:wire.HeaderSize])
						copy(newpkt[wire.HeaderSize:], ct)
					} else {
						newpkt = make([]byte, wire.HeaderSize+len(ct))
						copy(newpkt, out[i][:wire.HeaderSize])
						copy(newpkt[wire.HeaderSize:], ct)
					}
					// update header length
					if vh, err := wire.ViewHeader(newpkt); err == nil {
						vh.SetLength(uint16(len(ct)))
					}
					out[i] = newpkt
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
			}
			// return pooled buffers to pool to be reused
			if wpool != nil && len(pooledPtrs) > 0 {
				for _, p := range pooledPtrs {
					wpool.Put(p)
				}
			}
		}
	}
}
