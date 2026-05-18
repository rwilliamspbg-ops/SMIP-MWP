package crypto

import (
	"crypto/rand"
	"testing"
)

// BenchmarkEncryptInPlace measures the zero-copy encrypt hot path used by the
// AF_XDP forwarder when payload buffers have capacity for the AEAD tag.
func BenchmarkEncryptInPlace(b *testing.B) {
	combined := make([]byte, 64)
	if _, err := rand.Read(combined); err != nil {
		b.Fatalf("rand: %v", err)
	}
	sess, err := NewHybridSession(combined, []byte("bench-session"))
	if err != nil {
		b.Fatalf("new session: %v", err)
	}

	// typical MTU-size payload (approx)
	payload := make([]byte, 1400)
	for i := range payload {
		payload[i] = byte(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// allocate per-iteration to mimic independent frames and avoid reuse
		buf := make([]byte, len(payload), len(payload)+TagSize)
		copy(buf, payload)
		if err := sess.EncryptInPlace(buf, uint64(i)); err != nil {
			b.Fatalf("encrypt inplace: %v", err)
		}
	}
}

// BenchmarkDecryptInPlace measures the zero-copy decrypt hot path.
func BenchmarkDecryptInPlace(b *testing.B) {
	combined := make([]byte, 64)
	if _, err := rand.Read(combined); err != nil {
		b.Fatalf("rand: %v", err)
	}
	sess, err := NewHybridSession(combined, []byte("bench-session"))
	if err != nil {
		b.Fatalf("new session: %v", err)
	}

	payload := make([]byte, 1400)
	for i := range payload {
		payload[i] = byte(i)
	}

	// prepare a bounded ciphertext pool to avoid OOM when b.N is large
	// (long benchtime can make b.N huge). Use a circular buffer of prepared
	// ciphertexts and index into it during the timed loop.
	maxPrep := 100000 // cap preparations to 100k entries
	prepN := b.N
	if prepN <= 0 {
		prepN = 1
	}
	if prepN > maxPrep {
		prepN = maxPrep
	}
	cts := make([][]byte, prepN)
	nonces := make([]uint64, prepN)
	for i := 0; i < prepN; i++ {
		nonce := uint64(i)
		ct, err := sess.Encrypt(payload, nonce)
		if err != nil {
			b.Fatalf("prepare encrypt alloc: %v", err)
		}
		dst := make([]byte, len(ct))
		copy(dst, ct)
		cts[i] = dst
		nonces[i] = nonce
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		idx := i % len(cts)
		if _, err := sess.DecryptInPlace(cts[idx], nonces[idx]); err != nil {
			b.Fatalf("decrypt inplace: %v", err)
		}
	}
}
