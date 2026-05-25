// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright 2026 rwilliamspbg-ops
//
// This file is part of SMIP-MWP.

package crypto

import (
	"crypto/ecdh"
	"crypto/hmac"
	"crypto/mlkem"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"time"
)

// SessionState defines the valid states in the handshake lifecycle.
type SessionState string

const (
	StateUninitialized      SessionState = "UNINITIALIZED"
	StateAwaitingPeerPubkey SessionState = "AWAITING_PEER_PUBKEY"
	StateReadyForAuth       SessionState = "READY_FOR_AUTH"
	StateEstablished        SessionState = "ESTABLISHED"
	StateTimedOut           SessionState = "TIMED_OUT"
)

// HandshakeTimeout is the maximum time allowed for completing a handshake.
const HandshakeTimeout = 30 * time.Second

// MaxReplayWindow is the number of sequence numbers to track for replay detection.
const MaxReplayWindow = 64

// RetryBackoffMultiplier is used for exponential backoff on handshake failures.
const RetryBackoffMultiplier = 2.0

// MaxRetries is the maximum number of handshake retries before aborting.
const MaxRetries = 3

// x25519PubSize and mlkemPubSize are the wire sizes for the combined public key.
const (
	x25519PubSize = 32
	mlkemPubSize  = mlkem.EncapsulationKeySize768 // 1184 bytes
	mlkemCtSize   = mlkem.CiphertextSize768        // 1088 bytes
)

// HybridKEX holds ephemeral x25519 + ML-KEM-768 key material for one handshake.
// X25519 scalar multiplication uses crypto/ecdh (stdlib, Go 1.20+).
// ML-KEM-768 is provided by crypto/mlkem (stdlib, Go 1.24+).
type HybridKEX struct {
	x25519Priv *ecdh.PrivateKey
	x25519Pub  *ecdh.PublicKey

	mlkemKey *mlkem.DecapsulationKey768 // private — used by initiator to decapsulate
	mlkemPub *mlkem.EncapsulationKey768 // public — shared in PublicKey()
}

// NewHybridKEX initializes a hybrid key exchange instance with real cryptographic keys.
func NewHybridKEX() (*HybridKEX, error) {
	h := &HybridKEX{}

	// x25519 keypair via crypto/ecdh
	x25519Priv, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("kex: x25519 key gen: %w", err)
	}
	h.x25519Priv = x25519Priv
	h.x25519Pub  = x25519Priv.PublicKey()

	// ML-KEM-768 keypair
	key, err := mlkem.GenerateKey768()
	if err != nil {
		return nil, fmt.Errorf("kex: mlkem key gen: %w", err)
	}
	h.mlkemKey = key
	h.mlkemPub = key.EncapsulationKey()

	return h, nil
}

// PublicKey returns the combined initiator public key: x25519_pub (32) || mlkem_pub (1184).
func (h *HybridKEX) PublicKey() ([]byte, error) {
	if h.mlkemPub == nil {
		return nil, fmt.Errorf("HybridKEX: mlkem key not initialised")
	}
	pub := make([]byte, x25519PubSize+mlkemPubSize)
	copy(pub[:x25519PubSize], h.x25519Pub.Bytes())
	copy(pub[x25519PubSize:], h.mlkemPub.Bytes())
	return pub, nil
}

// Respond is called by the responder with the initiator's combined public key.
// It returns (responderMsg, sharedSecret, error).
//
// responderMsg wire format: x25519_resp_pub (32) || mlkem_ciphertext (1088)
func (h *HybridKEX) Respond(initiatorPub []byte) ([]byte, []byte, error) {
	if len(initiatorPub) < x25519PubSize+mlkemPubSize {
		return nil, nil, fmt.Errorf("kex: initiator pub too short: got %d, want %d",
			len(initiatorPub), x25519PubSize+mlkemPubSize)
	}

	// X25519 DH via crypto/ecdh
	peerX25519Pub, err := ecdh.X25519().NewPublicKey(initiatorPub[:x25519PubSize])
	if err != nil {
		return nil, nil, fmt.Errorf("kex: parse initiator x25519 pub: %w", err)
	}
	x25519SS, err := h.x25519Priv.ECDH(peerX25519Pub)
	if err != nil {
		return nil, nil, fmt.Errorf("kex: x25519 DH: %w", err)
	}

	// ML-KEM-768 encapsulation
	initiatorMLKEMPubBytes := initiatorPub[x25519PubSize : x25519PubSize+mlkemPubSize]
	initiatorMLKEMPub, err := mlkem.NewEncapsulationKey768(initiatorMLKEMPubBytes)
	if err != nil {
		return nil, nil, fmt.Errorf("kex: parse initiator mlkem pub: %w", err)
	}
	mlkemCiphertext, mlkemSS := initiatorMLKEMPub.Encapsulate()

	// Wire message: responder x25519 pub || mlkem ciphertext
	msg := make([]byte, x25519PubSize+mlkemCtSize)
	copy(msg[:x25519PubSize], h.x25519Pub.Bytes())
	copy(msg[x25519PubSize:], mlkemCiphertext)

	transcript := handshakeTranscript(initiatorPub[:x25519PubSize], h.x25519Pub.Bytes(), initiatorMLKEMPubBytes, mlkemCiphertext)
	ss, err := handshakeDeriveSecret(x25519SS, mlkemSS, transcript)
	if err != nil {
		return nil, nil, err
	}
	return msg, ss, nil
}

// Finish is called by the initiator upon receiving the responder's wire message.
// It decapsulates the ML-KEM ciphertext, performs X25519 DH, and returns the
// same 64-byte session secret produced by Respond.
func (h *HybridKEX) Finish(responderMsg []byte) ([]byte, error) {
	if len(responderMsg) < x25519PubSize+mlkemCtSize {
		return nil, fmt.Errorf("kex: responder msg too short: got %d, want %d",
			len(responderMsg), x25519PubSize+mlkemCtSize)
	}

	respX25519Pub, err := ecdh.X25519().NewPublicKey(responderMsg[:x25519PubSize])
	if err != nil {
		return nil, fmt.Errorf("kex: parse responder x25519 pub: %w", err)
	}
	x25519SS, err := h.x25519Priv.ECDH(respX25519Pub)
	if err != nil {
		return nil, fmt.Errorf("kex: x25519 DH: %w", err)
	}

	ct := responderMsg[x25519PubSize : x25519PubSize+mlkemCtSize]
	mlkemSS, err := h.mlkemKey.Decapsulate(ct)
	if err != nil {
		return nil, fmt.Errorf("kex: mlkem decapsulate: %w", err)
	}

	initiatorMLKEMPubBytes := h.mlkemPub.Bytes()
	transcript := handshakeTranscript(h.x25519Pub.Bytes(), respX25519Pub.Bytes(), initiatorMLKEMPubBytes, ct)
	return handshakeDeriveSecret(x25519SS, mlkemSS, transcript)
}

// handshakeTranscript builds a deterministic salt from the handshake transcript so
// both peers derive the same HKDF salt without any extra round-trip.
func handshakeTranscript(x25519InitPub, x25519RespPub, mlkemInitPub, mlkemCiphertext []byte) []byte {
	h := sha256.New()
	h.Write(x25519InitPub)
	h.Write(x25519RespPub)
	h.Write(mlkemInitPub)
	h.Write(mlkemCiphertext)
	return h.Sum(nil)
}

// handshakeDeriveSecret combines X25519 and ML-KEM shared secrets via two-stage HKDF.
func handshakeDeriveSecret(x25519SS, mlkemSS, transcript []byte) ([]byte, error) {
	prkClassical := handshakeHKDFExtract(transcript, x25519SS)
	prkPQC := handshakeHKDFExtract(transcript, mlkemSS)
	combined := append(prkClassical, prkPQC...)
	prkCombined := handshakeHKDFExtract(transcript, combined)

	// HKDF-Expand(prkCombined, info="smip-mwp-kex-v1", 64) — RFC 5869 §2.3
	info := []byte("smip-mwp-kex-v1")
	out := hkdfExpand(prkCombined, info, 64)
	return out, nil
}


// hkdfExpand computes HKDF-Expand(prk, info, length) via HMAC-SHA256 (RFC 5869 §2.3).
func hkdfExpand(prk, info []byte, length int) []byte {
	var out []byte
	var T []byte
	for i := 1; len(out) < length; i++ {
		mac := hmac.New(sha256.New, prk)
		mac.Write(T)
		mac.Write(info)
		mac.Write([]byte{byte(i)})
		T = mac.Sum(nil)
		out = append(out, T...)
	}
	return out[:length]
}

// handshakeHKDFExtract computes HKDF-Extract(salt, ikm) using stdlib HMAC-SHA256.
// This avoids the golang.org/x/crypto/hkdf dependency (RFC 5869 §2.2).
func handshakeHKDFExtract(salt, ikm []byte) []byte {
	if len(salt) == 0 {
		salt = make([]byte, sha256.Size)
	}
	mac := hmac.New(sha256.New, salt)
	mac.Write(ikm)
	return mac.Sum(nil)
}

// ---------------------------------------------------------------------------
// HybridKEXState — per-session handshake state machine
// ---------------------------------------------------------------------------

// HybridKEXState represents the state for a hybrid KEX handshake session.
type HybridKEXState struct {
	sessionID     [16]byte
	kexStarted    time.Time
	timeout       time.Time
	retryCount    int
	handshakeDone bool

	seqCounter uint64
	seqWindow  map[uint64]struct{}
}

// NewHybridKEXState creates a new handshake state.
func NewHybridKEXState(sessionID [16]byte) *HybridKEXState {
	return &HybridKEXState{
		sessionID:  sessionID,
		seqWindow:  make(map[uint64]struct{}),
		kexStarted: time.Now(),
		timeout:    time.Now().Add(HandshakeTimeout),
	}
}

// CheckTimeout checks if handshake has timed out.
func (s *HybridKEXState) CheckTimeout() error {
	if !s.handshakeDone {
		if time.Now().After(s.timeout) {
			return fmt.Errorf("crypto: handshake timeout for session %x", s.sessionID[:])
		}
		s.timeout = time.Now().Add(HandshakeTimeout)
	}
	return nil
}

// IncrementSeqCounter increments the sequence counter and checks the replay window.
func (s *HybridKEXState) IncrementSeqCounter() (uint64, error) {
	s.seqCounter++
	seq := s.seqCounter

	if _, exists := s.seqWindow[seq]; exists {
		return seq, fmt.Errorf("crypto: replay attack detected for session %x", s.sessionID[:])
	}

	// Evict oldest entry when window is full (O(n) scan is fine at MaxReplayWindow=64).
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

// CheckRetries returns an error if the max retry limit has been exceeded.
func (s *HybridKEXState) CheckRetries() error {
	if s.retryCount >= MaxRetries {
		return fmt.Errorf("crypto: handshake retry limit (%d) exceeded for session %x", MaxRetries, s.sessionID[:])
	}
	s.retryCount++
	return nil
}

// ResetRetry resets the retry counter on success.
func (s *HybridKEXState) ResetRetry() {
	s.retryCount = 0
}

// Cleanup zeros time fields and discards replay window state.
func (s *HybridKEXState) Cleanup() {
	s.kexStarted = time.Time{}
	s.timeout = time.Time{}
	s.seqWindow = make(map[uint64]struct{})
}
