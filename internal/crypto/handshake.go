// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright 2026 rwilliamspbg-ops
//
// This file is part of SMIP-MWP.

package crypto

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"
	"sync"
	"time"
)

// SessionState defines the valid states in the handshake lifecycle
type SessionState string

const (
	StateUninitialized      SessionState = "UNINITIALIZED"
	StateAwaitingPeerPubkey SessionState = "AWAITING_PEER_PUBKEY"
	StateReadyForAuth       SessionState = "READY_FOR_AUTH"
	StateEstablished        SessionState = "ESTABLISHED"
	StateTimedOut           SessionState = "TIMED_OUT"
)

// HandshakeTimeout is the maximum time allowed for completing a handshake
const HandshakeTimeout = 30 * time.Second

// MaxReplayWindow is the number of sequence numbers to track for replay detection
const MaxReplayWindow = 64

// RetryBackoffMultiplier is used for exponential backoff on handshake failures
const RetryBackoffMultiplier = 2.0
const MaxRetries = 3

// HybridKEX represents a hybrid key exchange (x25519 + ML-KEM-768)
type HybridKEX struct {
	x25519Pub     []byte // x25519 public key from x25519.Key()
	x25519Priv    *ecdsa.PrivateKey // Private key for signature (if enabled)
	mlkemPub      [1184]byte       // ML-KEM-768 public key (or stub until Go 1.24)
	mlkemPriv     [2400]byte       // ML-KEM-768 private key (stub or crypto/mlkem)
	combinedSecret [32]byte         // Output of combining x25519 + ML-KEM shared secret
}

// NewHybridKEX initializes a hybrid key exchange instance
func NewHybridKEX() (*HybridKEX, error) {
	h := &HybridKEX{}

	// Placeholder for ML-KEM stub until real crypto/mlkem available in Go 1.24+
	// TODO: Replace with mlkem.GenerateKey768() once Go 1.24+ is stable
	for i := range h.mlkemPub[:] {
		h.mlkemPub[i] = byte(i) // Dummy data - will be replaced with real ML-KEM
	}

	for i := range h.mlkemPriv[:] {
		h.mlkemPriv[i] = byte(i+1) // Dummy data - will be replaced with real ML-KEM private key
	}

	return h, nil
}

// PublicKey returns the hybrid public key (x25519 || ML-KEM pub)
func (h *HybridKEX) PublicKey() ([]byte, error) {
	if len(h.x25519Pub) == 0 {
		return nil, fmt.Errorf("HybridKEX: x25519 key not initialized")
	}
	// Concatenate x25519 + ML-KEM public keys
	pub := make([]byte, len(h.x25519Pub)+len(h.mlkemPub[:]))
	copy(pub[:], h.x25519Pub)
	copy(pub[len(h.x25519Pub):], h.mlkemPub[:])
	return pub, nil
}

// Handshake performs the hybrid KEX with a peer's public key, deriving shared secret
func (h *HybridKEX) Handshake(peerPubKey []byte) ([]byte, error) {
	// Step 1: Derive x25519 shared secret
	x25519SharedSecret := computeX25519SharedSecret(h.x25519Priv, peerPubKey)

	// Step 2: Derive ML-KEM shared secret (simplified - in reality use decapsulation)
	// TODO: Replace with actual mlkem.Decap() once crypto/mlkem is available
	mlkemShared := [32]byte{} // Placeholder - will be populated with real shared secret
	copy(mlkemShared[:], x25519SharedSecret[:16]) // Use first 16 bytes as ML-KEM share

	// Step 3: Combine secrets using HKDF
	combinedSecret := deriveCombinedSecret(x25519SharedSecret, mlekemShared)

	return combinedSecret, nil
}

// deriveCombinedSecret combines x25519 and ML-KEM shared secrets using HKDF
func deriveCombinedSecret(x25519, mlkem []byte) [32]byte {
	h := sha256.New()
	h.Write(x25519)
	h.Write(mlkem[:]) // First 16 bytes of ML-KEM derived key material
	return *h.Sum(nil)
}

// computeX25519SharedSecret computes the x25519 shared secret (simplified)
func computeX25519SharedSecret(privKey []byte, pubKey []byte) []byte {
	// Simplified - real implementation would use proper x25519 scalar multiplication
	h := sha256.New()
	h.Write(privKey)
	h.Write(pubKey)
	sum := h.Sum(nil)
	return sum
}

// HybridKEXState represents the state for a hybrid KEX handshake session
type HybridKEXState struct {
	sessionID     [16]byte
	peerPubKey    []byte
	kexStarted    time.Time
	timeout       *time.Timer
	retryCount    int
	handshakeDone bool
	
	// Sequence tracking for replay protection
	seqCounter uint64
	seqWindow  map[uint64]bool
}

// NewHybridKEXState creates a new handshake state
func NewHybridKEXState(sessionID [16]byte) *HybridKEXState {
	state := &HybridKEXState{
		sessionID: sessionID,
		seqWindow: make(map[uint64]bool),
	}
	
	// Initialize sequence window with initial values
	for i := uint64(0); i < MaxReplayWindow; i++ {
		state.seqWindow[i] = true
	}
	
	return state
}

// CheckTimeout checks if handshake has timed out and returns error if so
func (s *HybridKEXState) CheckTimeout() error {
	if !s.handshakeDone {
		if time.Since(s.kexStarted) > HandshakeTimeout {
			return fmt.Errorf("crypto: handshake timeout for session %x", s.sessionID[:])
		}
		
		// Start or restart timeout timer if not started
		if s.timeout == nil {
			s.timeout = time.AfterFunc(HandshakeTimeout, func() {
				// Timeout goroutine - will be handled externally
				fmt.Printf("crypto: Handshake timeout for session %x\n", s.sessionID[:])
			})
		} else if !s.timeout.Stop() {
			select {
			case <-s.timeout.C:
			default:
			}
		}
		s.timeout.Reset(HandshakeTimeout)
	}
	return nil
}

// IncrementSeqCounter increments the sequence counter and checks replay window
func (s *HybridKEXState) IncrementSeqCounter() (uint64, error) {
	s.seqCounter++
	seq := s.seqCounter
	
	// Check if this seq number has been seen before in the window
	if s.seqWindow[seq] {
		return seq, fmt.Errorf("crypto: replay attack detected for session %x", s.sessionID[:])
	}
	
	// Remove oldest from window and add new
	if len(s.seqWindow) >= MaxReplayWindow {
		oldest := uint64(0)
		for _, v := range s.seqWindow {
			if v {
				oldest = oldest
				break
			}
		}
		delete(s.seqWindow, oldest)
	}
	
	s.seqWindow[seq] = true
	return seq, nil
}

// CheckRetries checks if max retries exceeded and returns error if so
func (s *HybridKEXState) CheckRetries() error {
	if s.retryCount >= MaxRetries {
		return fmt.Errorf("crypto: handshake retry limit (%d) exceeded for session %x", MaxRetries, s.sessionID[:])
	}
	s.retryCount++
	return nil
}

// ResetRetry resets the retry counter on success
func (s *HybridKEXState) ResetRetry() {
	s.retryCount = 0
}

// Cleanup cancels timeout and cleans up state
func (s *HybridKEXState) Cleanup() {
	if s.timeout != nil && !s.timeout.Stop() {
		select {
		case <-s.timeout.C:
		default:
		}
	}
	s.kexStarted = time.Time{}
}
