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

	// prepare a ciphertext for each iteration using the allocation-based helper
	// which avoids potential aliasing issues when preparing ciphertexts in bulk.
	cts := make([][]byte, b.N)
	for i := 0; i < b.N; i++ {
		ct, err := sess.Encrypt(payload, uint64(i))
		if err != nil {
			b.Fatalf("prepare encrypt alloc: %v", err)
		}
		// copy to ensure each entry is independent
		dst := make([]byte, len(ct))
		copy(dst, ct)
		cts[i] = dst
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := sess.DecryptInPlace(cts[i], uint64(i)); err != nil {
			b.Fatalf("decrypt inplace: %v", err)
		}
	}
}
