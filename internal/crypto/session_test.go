package crypto

import (
	"crypto/rand"
	"testing"
)

func TestHybridSession_EncryptDecrypt_RoundTrip(t *testing.T) {
	// Create a random combined secret (64 bytes) and session info
	combined := make([]byte, 64)
	if _, err := rand.Read(combined); err != nil {
		t.Fatalf("rand: %v", err)
	}
	sessionInfo := []byte("test-session-info")

	a, err := NewHybridSession(combined, sessionInfo)
	if err != nil {
		t.Fatalf("new session A: %v", err)
	}
	b, err := NewHybridSession(combined, sessionInfo)
	if err != nil {
		t.Fatalf("new session B: %v", err)
	}
	// For test purposes, align nonce base and seqMask so both sessions derive identical nonces.
	b.nonceBase = a.nonceBase
	b.seqMask = a.seqMask

	plaintext := []byte("hello hybrid session world")

	// Slow-path encrypt (allocating) and decrypt on the peer
	ct, err := a.Encrypt(plaintext, 1)
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	pt, err := b.Decrypt(ct, 1)
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}
	if string(pt) != string(plaintext) {
		t.Fatalf("roundtrip mismatch: got %q want %q", string(pt), string(plaintext))
	}
}

func TestHybridSession_EncryptInPlace_DecryptInPlace(t *testing.T) {
	combined := make([]byte, 64)
	if _, err := rand.Read(combined); err != nil {
		t.Fatalf("rand: %v", err)
	}
	sessionInfo := []byte("inplace-test")

	s, err := NewHybridSession(combined, sessionInfo)
	if err != nil {
		t.Fatalf("new session: %v", err)
	}

	plaintext := []byte("inline payload for encryption")
	// prepare buffer with extra capacity for tag (TagSize)
	buf := make([]byte, len(plaintext), len(plaintext)+TagSize)
	copy(buf, plaintext)

	if err := s.EncryptInPlace(buf[:len(plaintext)], 7); err != nil {
		t.Fatalf("encrypt in place: %v", err)
	}
	// Caller must reslice to include appended tag.
	buf = buf[:len(plaintext)+TagSize]

	pt, err := s.DecryptInPlace(buf, 7)
	if err != nil {
		t.Fatalf("decrypt in place: %v", err)
	}
	if string(pt) != string(plaintext) {
		t.Fatalf("inplace roundtrip mismatch: got %q want %q", string(pt), string(plaintext))
	}
}
