// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright 2026 rwilliamspbg-ops
//
// This file is part of SMIP-MWP.
// SMIP-MWP is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation; either version 3 of the License, or (at your option) any later version.
// See the LICENSE file in the project root for details.

package crypto

import (
	"crypto/mlkem"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"hash"
	"io"

	"golang.org/x/crypto/curve25519"
	"golang.org/x/crypto/hkdf"
)

// HybridKeyExchange performs x25519 + ML-KEM-768 hybrid key exchange.
//
// Protocol (two-phase, asymmetric):
//
//	Initiator                              Responder
//	─────────────────────────────────────────────────
//	NewHybridKEX() → kex
//	kex.PublicKey() ──────────────────────▶ initiatorPub
//	                                        NewHybridKEX() → kex
//	                                        kex.Respond(initiatorPub)
//	                                          → (responderMsg, sharedSecret)
//	responderMsg   ◀──────────────────────  send responderMsg
//	kex.Finish(responderMsg)
//	  → sharedSecret
//
// Both sides derive the same 64-byte session secret.
//
// Key derivation:
//
//	transcript_salt = SHA-256(x25519_init_pub || x25519_resp_pub || mlkem_init_pub || mlkem_ciphertext)
//	prk_classical   = HKDF-Extract(transcript_salt, x25519_ss)
//	prk_pqc         = HKDF-Extract(transcript_salt, mlkem_ss)
//	combined_prk    = HKDF-Extract(transcript_salt, prk_classical || prk_pqc)
//	session_secret  = HKDF-Expand(combined_prk, "smip-mwp-kex-v1", 64)
type HybridKeyExchange struct {
	x25519Priv [32]byte
	x25519Pub  [32]byte

	// ML-KEM-768 keypair — used for decapsulation by the initiator.
	mlkemKey *mlkem.DecapsulationKey768
	mlkemPub *mlkem.EncapsulationKey768
}

// NewHybridKEX generates a fresh x25519 + ML-KEM-768 ephemeral keypair.
func NewHybridKEX(rng io.Reader) (*HybridKeyExchange, error) {
	if rng == nil {
		rng = rand.Reader
	}
	h := &HybridKeyExchange{}

	if _, err := io.ReadFull(rng, h.x25519Priv[:]); err != nil {
		return nil, fmt.Errorf("kex: x25519 key gen: %w", err)
	}
	pub, err := curve25519.X25519(h.x25519Priv[:], curve25519.Basepoint)
	if err != nil {
		return nil, fmt.Errorf("kex: x25519 public key: %w", err)
	}
	copy(h.x25519Pub[:], pub)

	key, err := mlkem.GenerateKey768()
	if err != nil {
		return nil, fmt.Errorf("kex: mlkem key gen: %w", err)
	}
	h.mlkemKey = key
	h.mlkemPub = key.EncapsulationKey()

	return h, nil
}

// PublicKey returns the initiator's combined public key: x25519_pub (32) || mlkem_pub (1184).
// Send this to the responder to begin the handshake.
func (h *HybridKeyExchange) PublicKey() []byte {
	mlkemPubBytes := h.mlkemPub.Bytes()
	out := make([]byte, 32+len(mlkemPubBytes))
	copy(out[:32], h.x25519Pub[:])
	copy(out[32:], mlkemPubBytes)
	return out
}

// ResponderMessage is returned by Respond and contains the responder's
// x25519 public key (32 bytes) plus the ML-KEM-768 ciphertext (1088 bytes).
// Total wire size: 1120 bytes.
type ResponderMessage struct {
	X25519Pub      [32]byte
	MLKEMCiphertext [mlkem.CiphertextSize768]byte
}

// Bytes serialises the responder message to wire format: x25519_pub || mlkem_ciphertext.
func (m *ResponderMessage) Bytes() []byte {
	out := make([]byte, 32+mlkem.CiphertextSize768)
	copy(out[:32], m.X25519Pub[:])
	copy(out[32:], m.MLKEMCiphertext[:])
	return out
}

// Respond is called by the responder upon receiving the initiator's public key.
// It encapsulates against the initiator's ML-KEM public key, performs X25519 DH,
// and returns the responder's wire message together with the derived 64-byte session secret.
func (h *HybridKeyExchange) Respond(initiatorPub []byte) (*ResponderMessage, []byte, error) {
	if len(initiatorPub) < 32+mlkem.EncapsulationKeySize768 {
		return nil, nil, fmt.Errorf("kex: initiator public key too short: got %d, want %d",
			len(initiatorPub), 32+mlkem.EncapsulationKeySize768)
	}

	// --- Classical: X25519 ---
	var peerX25519 [32]byte
	copy(peerX25519[:], initiatorPub[:32])
	x25519SS, err := curve25519.X25519(h.x25519Priv[:], peerX25519[:])
	if err != nil {
		return nil, nil, fmt.Errorf("kex: x25519 DH: %w", err)
	}

	// --- PQC: ML-KEM-768 encapsulation ---
	initiatorMLKEMPubBytes := initiatorPub[32 : 32+mlkem.EncapsulationKeySize768]
	initiatorMLKEMPub, err := mlkem.NewEncapsulationKey768(initiatorMLKEMPubBytes)
	if err != nil {
		return nil, nil, fmt.Errorf("kex: parse initiator mlkem public key: %w", err)
	}
	mlkemCiphertext, mlkemSS := initiatorMLKEMPub.Encapsulate()

	msg := &ResponderMessage{}
	copy(msg.X25519Pub[:], h.x25519Pub[:])
	copy(msg.MLKEMCiphertext[:], mlkemCiphertext)

	// Derive shared secret
	transcript := buildTranscript(peerX25519[:], h.x25519Pub[:], initiatorMLKEMPubBytes, mlkemCiphertext)
	sessionSecret, err := deriveSessionSecret(x25519SS, mlkemSS, transcript)
	if err != nil {
		return nil, nil, err
	}

	return msg, sessionSecret, nil
}

// Finish is called by the initiator upon receiving the responder's message.
// It decapsulates the ML-KEM ciphertext, performs X25519 DH, and returns
// the same 64-byte session secret that Respond produced.
func (h *HybridKeyExchange) Finish(responderMsg []byte) ([]byte, error) {
	if len(responderMsg) < 32+mlkem.CiphertextSize768 {
		return nil, fmt.Errorf("kex: responder message too short: got %d, want %d",
			len(responderMsg), 32+mlkem.CiphertextSize768)
	}

	// --- Classical: X25519 ---
	var respX25519Pub [32]byte
	copy(respX25519Pub[:], responderMsg[:32])
	x25519SS, err := curve25519.X25519(h.x25519Priv[:], respX25519Pub[:])
	if err != nil {
		return nil, fmt.Errorf("kex: x25519 DH: %w", err)
	}

	// --- PQC: ML-KEM-768 decapsulation ---
	var ctBytes [mlkem.CiphertextSize768]byte
	copy(ctBytes[:], responderMsg[32:32+mlkem.CiphertextSize768])
	mlkemSS, err := h.mlkemKey.Decapsulate(ctBytes[:])
	if err != nil {
		return nil, fmt.Errorf("kex: mlkem decapsulate: %w", err)
	}

	// Rebuild transcript using initiator's own mlkem pub + received ciphertext
	initiatorMLKEMPubBytes := h.mlkemPub.Bytes()
	transcript := buildTranscript(h.x25519Pub[:], respX25519Pub[:], initiatorMLKEMPubBytes, ctBytes[:])

	return deriveSessionSecret(x25519SS, mlkemSS, transcript)
}

// buildTranscript constructs the HKDF salt as a hash of the handshake transcript,
// ensuring both peers use the same salt without any additional round-trip.
//
//	transcript = SHA-256(x25519_init_pub || x25519_resp_pub || mlkem_init_pub || mlkem_ciphertext)
func buildTranscript(x25519InitPub, x25519RespPub, mlkemInitPub, mlkemCiphertext []byte) []byte {
	h := sha256.New()
	h.Write(x25519InitPub)
	h.Write(x25519RespPub)
	h.Write(mlkemInitPub)
	h.Write(mlkemCiphertext)
	return h.Sum(nil)
}

// deriveSessionSecret combines classical and PQC shared secrets via two-stage HKDF.
//
//	prk_classical  = HKDF-Extract(transcript, x25519_ss)
//	prk_pqc        = HKDF-Extract(transcript, mlkem_ss)
//	combined_prk   = HKDF-Extract(transcript, prk_classical || prk_pqc)
//	session_secret = HKDF-Expand(combined_prk, "smip-mwp-kex-v1", 64)
func deriveSessionSecret(x25519SS, mlkemSS, transcript []byte) ([]byte, error) {
	prkClassical := hkdfExtract(sha256.New, transcript, x25519SS)
	prkPQC := hkdfExtract(sha256.New, transcript, mlkemSS)

	combined := append(prkClassical, prkPQC...)
	prkCombined := hkdfExtract(sha256.New, transcript, combined)

	r := hkdf.New(sha256.New, prkCombined, transcript, []byte("smip-mwp-kex-v1"))
	sessionSecret := make([]byte, 64)
	if _, err := io.ReadFull(r, sessionSecret); err != nil {
		return nil, fmt.Errorf("kex: session secret expand: %w", err)
	}
	return sessionSecret, nil
}

// hkdfExtract is HKDF-Extract(salt, ikm) → prk.
func hkdfExtract(h func() hash.Hash, salt, ikm []byte) []byte {
	r := hkdf.New(h, ikm, salt, nil)
	prk := make([]byte, sha256.Size)
	_, _ = io.ReadFull(r, prk)
	return prk
}
