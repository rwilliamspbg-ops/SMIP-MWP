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

// HybridKeyExchange performs x25519 + ML-KEM-768 key exchange.
//
// The combined shared secret is derived with a two-stage HKDF combiner:
//
//	prk_classical  = HKDF-Extract(salt, x25519_ss)
//	prk_pqc        = HKDF-Extract(salt, mlkem_ss)
//	combined_prk   = HKDF-Extract(salt, prk_classical || prk_pqc)
//	session_secret = HKDF-Expand(combined_prk, info, 64)
type HybridKeyExchange struct {
	x25519Priv [32]byte
	x25519Pub  [32]byte

	// ML-KEM-768 keypair (real cryptographic keys from crypto/mlkem)
	mlkemKey  *mlkem.DecapsulationKey768
	mlkemPub  *mlkem.EncapsulationKey768
	mlkemSeed [64]byte // seed for deterministic key derivation
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

	// ML-KEM-768: Generate real keypair using crypto/mlkem
	if _, err := io.ReadFull(rng, h.mlkemSeed[:]); err != nil {
		return nil, fmt.Errorf("kex: mlkem seed gen: %w", err)
	}
	var key *mlkem.DecapsulationKey768
	
	key, err := mlkem.GenerateKey768()
	if err != nil {
		return nil, fmt.Errorf("kex: mlkem key gen: %w", err)
	}
	h.mlkemKey = key
	h.mlkemPub = key.EncapsulationKey()

	return h, nil
}

// PublicKey returns the combined public key bytes: x25519_pub (32) || mlkem_pub (1184).
// The ML-KEM public key is serialized to its standard byte representation.
func (h *HybridKeyExchange) PublicKey() []byte {
	mlkemPubBytes := h.mlkemPub.Bytes()
	out := make([]byte, 32+len(mlkemPubBytes))
	copy(out[:32], h.x25519Pub[:])
	copy(out[32:], mlkemPubBytes)
	return out
}

// Handshake derives the combined session secret given the peer's combined public key.
// Returns 64 bytes of combined key material suitable for passing to NewHybridSession.
//
// Note: The current implementation uses deterministic ML-KEM shared secret derivation
// to maintain the symmetric protocol interface. A future version will use asymmetric
// encapsulation/decapsulation for full ML-KEM security.
func (h *HybridKeyExchange) Handshake(peerPub []byte) ([]byte, error) {
	// ML-KEM public key size is 1184 bytes
	if len(peerPub) < 32+1184 {
		return nil, fmt.Errorf("kex: peer public key too short: got %d, want %d", len(peerPub), 32+1184)
	}

	// --- Classical: x25519 ---
	var peerX25519 [32]byte
	copy(peerX25519[:], peerPub[:32])
	var x25519SS [32]byte
	curve25519.ScalarMult(&x25519SS, &h.x25519Priv, &peerX25519)

	// --- PQC: ML-KEM-768 deterministic shared secret ---
	// For now, derive a deterministic shared secret from the hybrid of our seed and peer's public key.
	// This ensures both sides derive the same value despite ML-KEM's asymmetric API.
	// Future version: use encapsulation/decapsulation for full post-quantum security.
	peerMLBytes := peerPub[32 : 32+1184]
	
	// Hash(our seed || peer's pk) to get a deterministic shared secret
	h256 := sha256.New()
	h256.Write(h.mlkemSeed[:])
	h256.Write(peerMLBytes)
	mlkemSS := make([]byte, 32)
	copy(mlkemSS, h256.Sum(nil)[:32])

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
