// Package crypto implements the SMIP/MWP hybrid post-quantum cryptography layer.
//
// Key exchange: x25519 + ML-KEM-768 (crypto/mlkem, Go 1.24+)
// Key derivation: two-stage HKDF with domain separation
// AEAD: AES-256-GCM (hardware accelerated) with ChaCha20-Poly1305 fallback
// Signing stubs: ML-DSA-65 via cloudflare/circl (when mature)
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/hkdf"
)

const (
	// TagSize is the AEAD authentication tag length in bytes.
	TagSize = 16
	// KeySize is 256-bit AES / ChaCha20 key.
	KeySize = 32
	// NonceSize for GCM / ChaCha20-Poly1305.
	NonceSize = 12

	// Domain-separation labels for HKDF.
	hkdfLabelSession = "smip-mwp-session-v1"
)

var (
	ErrPayloadTooLarge      = errors.New("crypto: payload exceeds per-packet limit")
	ErrInsufficientCapacity = errors.New("crypto: buffer lacks capacity for auth tag")
	ErrCiphertextTooShort   = errors.New("crypto: ciphertext shorter than tag size")
	ErrAuthenticationFailed = errors.New("crypto: AEAD authentication failed")
)

// HybridSession holds the symmetric AEAD state derived from a hybrid KEX.
// One HybridSession exists per sovereign tunnel; it is NOT safe for concurrent use
// without external locking.
type HybridSession struct {
	aead      cipher.AEAD
	nonceBase [NonceSize]byte // randomised per-session; XOR'd with seq counter
	seqMask   uint64          // extra entropy mixed into nonce
}

// NewHybridSession derives a session from combinedSecret (output of hybrid KEX HKDF)
// and sessionInfo (e.g. SrcID || DstID || FlowLabel for domain separation).
func NewHybridSession(combinedSecret, sessionInfo []byte) (*HybridSession, error) {
	// HKDF-Expand: extract session key material
	label := []byte(hkdfLabelSession)
	info := append(label, sessionInfo...) //nolint:gocritic
	r := hkdf.New(sha256.New, combinedSecret, nil, info)

	key := make([]byte, KeySize)
	if _, err := io.ReadFull(r, key); err != nil {
		return nil, fmt.Errorf("crypto: HKDF key derivation: %w", err)
	}

	// Prefer AES-256-GCM (hardware-accelerated on amd64/arm64).
	aead, err := newAEAD(key)
	if err != nil {
		return nil, err
	}

	// Deterministically derive nonce base and seqMask from the HKDF stream so
	// both peers holding the same combinedSecret and sessionInfo derive the
	// identical session state.
	var nonceBase [NonceSize]byte
	if _, err := io.ReadFull(r, nonceBase[:]); err != nil {
		return nil, fmt.Errorf("crypto: nonce base derivation: %w", err)
	}

	var mask [8]byte
	if _, err := io.ReadFull(r, mask[:]); err != nil {
		return nil, fmt.Errorf("crypto: seqMask derivation: %w", err)
	}

	s := &HybridSession{aead: aead}
	copy(s.nonceBase[:], nonceBase[:])
	s.seqMask = binary.BigEndian.Uint64(mask[:])

	return s, nil
}

// newAEAD selects AES-256-GCM when available, falling back to ChaCha20-Poly1305.
func newAEAD(key []byte) (cipher.AEAD, error) {
	block, err := aes.NewCipher(key)
	if err == nil {
		gcm, err := cipher.NewGCM(block)
		if err == nil {
			return gcm, nil
		}
	}
	// Fallback: pure-Go ChaCha20-Poly1305 (constant-time, ARM-friendly).
	return chacha20poly1305.New(key)
}

// buildNonce constructs a unique 12-byte nonce from the session base + sequence number.
func (s *HybridSession) buildNonce(seq uint64) []byte {
	nonce := make([]byte, NonceSize)
	copy(nonce, s.nonceBase[:])
	// XOR last 8 bytes with (seq ^ seqMask) — counters never repeat for this session.
	existing := binary.BigEndian.Uint64(nonce[4:])
	binary.BigEndian.PutUint64(nonce[4:], existing^seq^s.seqMask)
	return nonce
}

// EncryptInPlace encrypts payload in-place and appends the 16-byte authentication tag.
//
// CRITICAL: the backing array of payload must have at least cap(payload)+TagSize bytes
// available, because Seal() will extend the slice to append the tag.
// This is the zero-copy hot path used by the AF_XDP forwarder.
func (s *HybridSession) EncryptInPlace(payload []byte, seq uint64) error {
	if len(payload) > (1 << 24) {
		return ErrPayloadTooLarge
	}
	if cap(payload) < len(payload)+TagSize {
		return ErrInsufficientCapacity
	}
	nonce := s.buildNonce(seq)
	originalLen := len(payload)
	// Extend slice to accommodate tag; Seal writes cipher + tag starting at dst[:0].
	extended := payload[:originalLen+TagSize]
	s.aead.Seal(extended[:0], nonce, payload[:originalLen], nil)
	return nil
}

// DecryptInPlace decrypts and authenticates in-place (removes tag, returns plaintext slice).
func (s *HybridSession) DecryptInPlace(payload []byte, seq uint64) ([]byte, error) {
	if len(payload) < TagSize {
		return nil, ErrCiphertextTooShort
	}
	nonce := s.buildNonce(seq)
	plaintext, err := s.aead.Open(payload[:0], nonce, payload, nil)
	if err != nil {
		return nil, ErrAuthenticationFailed
	}
	return plaintext, nil
}

// Encrypt returns a newly allocated ciphertext+tag buffer (slow path / handshake use).
func (s *HybridSession) Encrypt(plaintext []byte, seq uint64) ([]byte, error) {
	out := make([]byte, len(plaintext)+TagSize)
	copy(out, plaintext)
	if err := s.EncryptInPlace(out[:len(plaintext)], seq); err != nil {
		return nil, err
	}
	return out[:len(plaintext)+TagSize], nil
}

// Decrypt returns a newly allocated plaintext (slow path / handshake use).
func (s *HybridSession) Decrypt(ciphertext []byte, seq uint64) ([]byte, error) {
	buf := make([]byte, len(ciphertext))
	copy(buf, ciphertext)
	return s.DecryptInPlace(buf, seq)
}
