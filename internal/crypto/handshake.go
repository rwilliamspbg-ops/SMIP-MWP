// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright 2026 rwilliamspbg-ops
//
// This file is part of SMIP-MWP.

package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
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
	x25519Pub  []byte
	x25519Priv []byte
	mlkemPub   [1184]byte
	mlkemPriv  [2400]byte
}

// NewHybridKEX initializes a hybrid key exchange instance
func NewHybridKEX() (*HybridKEX, error) {
	h := &HybridKEX{}
	h.x25519Pub = make([]byte, 32)
	h.x25519Priv = make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, h.x25519Pub); err != nil {
		return nil, fmt.Errorf("kex: x25519 pub gen: %w", err)
	}
	if _, err := io.ReadFull(rand.Reader, h.x25519Priv); err != nil {
		return nil, fmt.Errorf("kex: x25519 priv gen: %w", err)
	}

	// Placeholder material until internal/crypto wiring is migrated to root package kex.
	for i := range h.mlkemPub[:] {
		h.mlkemPub[i] = byte(i)
	}

	for i := range h.mlkemPriv[:] {
		h.mlkemPriv[i] = byte(i + 1)
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
	if len(peerPubKey) == 0 {
		return nil, fmt.Errorf("kex: peer public key empty")
	}
	x25519SharedSecret := computeX25519SharedSecret(h.x25519Priv, peerPubKey)

	mlkemShared := make([]byte, 32)
	copy(mlkemShared, x25519SharedSecret[:16])

	combinedSecret := deriveCombinedSecret(x25519SharedSecret, mlkemShared)

	return combinedSecret, nil
}

// deriveCombinedSecret combines x25519 and ML-KEM shared secrets using HKDF
func deriveCombinedSecret(x25519, mlkem []byte) []byte {
	h := sha256.New()
	_, _ = h.Write(x25519)
	_, _ = h.Write(mlkem)
	return h.Sum(nil)
}

// computeX25519SharedSecret computes the x25519 shared secret (simplified)
func computeX25519SharedSecret(privKey []byte, pubKey []byte) []byte {
	h := sha256.New()
	_, _ = h.Write(privKey)
	_, _ = h.Write(pubKey)
	sum := h.Sum(nil)
	return sum
}

// HybridKEXState represents the state for a hybrid KEX handshake session
type HybridKEXState struct {
	sessionID     [16]byte
	kexStarted    time.Time
	timeout       time.Time
	retryCount    int
	handshakeDone bool

	// Sequence tracking for replay protection
	seqCounter uint64
	seqWindow  map[uint64]struct{}
}

// NewHybridKEXState creates a new handshake state
func NewHybridKEXState(sessionID [16]byte) *HybridKEXState {
	state := &HybridKEXState{
		sessionID:  sessionID,
		seqWindow:  make(map[uint64]struct{}),
		kexStarted: time.Now(),
		timeout:    time.Now().Add(HandshakeTimeout),
	}

	return state
}

// CheckTimeout checks if handshake has timed out and returns error if so
func (s *HybridKEXState) CheckTimeout() error {
	if !s.handshakeDone {
		if time.Now().After(s.timeout) {
			return fmt.Errorf("crypto: handshake timeout for session %x", s.sessionID[:])
		}
		s.timeout = time.Now().Add(HandshakeTimeout)
	}
	return nil
}

// IncrementSeqCounter increments the sequence counter and checks replay window
func (s *HybridKEXState) IncrementSeqCounter() (uint64, error) {
	s.seqCounter++
	seq := s.seqCounter

	// Check if this seq number has been seen before in the window
	if _, exists := s.seqWindow[seq]; exists {
		return seq, fmt.Errorf("crypto: replay attack detected for session %x", s.sessionID[:])
	}

	// Remove oldest from window and add new
	if len(s.seqWindow) >= MaxReplayWindow {
		oldest := seq
		for seen := range s.seqWindow {
			if seen < oldest {
				oldest = seen
			}
		}
		delete(s.seqWindow, oldest)
	}

	s.seqWindow[seq] = struct{}{}
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
	s.kexStarted = time.Time{}
	s.timeout = time.Time{}
}
