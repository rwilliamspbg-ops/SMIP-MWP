package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"hash"
	"io"

	"golang.org/x/crypto/curve25519"
	"golang.org/x/crypto/hkdf"
)

// HybridKeyExchange performs x25519 + ML-KEM-768 (stub) key exchange.
//
// The combined shared secret is derived with a two-stage HKDF combiner:
//
//	prk_classical  = HKDF-Extract(salt, x25519_ss)
//	prk_pqc        = HKDF-Extract(salt, mlkem_ss)   // stub: random bytes until Go stdlib matures
//	combined_prk   = HKDF-Extract(salt, prk_classical || prk_pqc)
//	session_secret = HKDF-Expand(combined_prk, info, 64)
type HybridKeyExchange struct {
	x25519Priv [32]byte
	x25519Pub  [32]byte

	// ML-KEM-768 public key material (stub — replace with crypto/mlkem once available)
	mlkemPub  []byte
	mlkemPriv []byte
}

// NewHybridKEX generates a fresh x25519 + ML-KEM-768 ephemeral keypair.
func NewHybridKEX(rng io.Reader) (*HybridKeyExchange, error) {
	if rng == nil {
		rng = rand.Reader
	}
	h := &HybridKeyExchange{}

	// x25519 private key
	if _, err := io.ReadFull(rng, h.x25519Priv[:]); err != nil {
		return nil, fmt.Errorf("kex: x25519 key gen: %w", err)
	}
	curve25519.ScalarBaseMult(&h.x25519Pub, &h.x25519Priv)

	// ML-KEM-768 stub: generate random 1184-byte "public key" placeholder.
	// TODO: replace with crypto/mlkem KEM once Go 1.24+ stdlib API stabilises.
	h.mlkemPub = make([]byte, 1184)
	h.mlkemPriv = make([]byte, 2400)
	if _, err := io.ReadFull(rng, h.mlkemPub); err != nil {
		return nil, fmt.Errorf("kex: mlkem pub gen: %w", err)
	}
	if _, err := io.ReadFull(rng, h.mlkemPriv); err != nil {
		return nil, fmt.Errorf("kex: mlkem priv gen: %w", err)
	}

	return h, nil
}

// PublicKey returns the combined public key bytes: x25519_pub (32) || mlkem_pub (1184).
func (h *HybridKeyExchange) PublicKey() []byte {
	out := make([]byte, 32+len(h.mlkemPub))
	copy(out[:32], h.x25519Pub[:])
	copy(out[32:], h.mlkemPub)
	return out
}

// Handshake derives the combined session secret given the peer's combined public key.
// Returns 64 bytes of combined key material suitable for passing to NewHybridSession.
func (h *HybridKeyExchange) Handshake(peerPub []byte) ([]byte, error) {
	if len(peerPub) < 32+1184 {
		return nil, fmt.Errorf("kex: peer public key too short")
	}

	// --- Classical: x25519 ---
	var peerX25519 [32]byte
	copy(peerX25519[:], peerPub[:32])
	var x25519SS [32]byte
	curve25519.ScalarMult(&x25519SS, &h.x25519Priv, &peerX25519)

	// --- PQC: ML-KEM-768 stub (random shared secret placeholder) ---
	// TODO: replace with actual KEM decapsulation once library is stable.
	mlkemSS := make([]byte, 32)
	if _, err := rand.Read(mlkemSS); err != nil {
		return nil, fmt.Errorf("kex: mlkem ss stub: %w", err)
	}

	// --- Two-stage HKDF combiner ---
	salt := make([]byte, 32)
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("kex: salt gen: %w", err)
	}

	prkClassical := hkdfExtract(sha256.New, salt, x25519SS[:])
	prkPQC := hkdfExtract(sha256.New, salt, mlkemSS)

	combined := append(prkClassical, prkPQC...)
	prkCombined := hkdfExtract(sha256.New, salt, combined)

	r := hkdf.New(sha256.New, prkCombined, salt, []byte("smip-mwp-kex-v1"))
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
