package crypto

import (
	"crypto/rand"
	"testing"
)

// BenchmarkNewHybridSession_Cached measures repeated NewHybridSession calls
// that should hit the HKDF cache.
func BenchmarkNewHybridSession_Cached(b *testing.B) {
	combined := make([]byte, 32)
	rand.Read(combined)
	sessionInfo := []byte("bench-session-cached")
	// warm cache
	_, _ = NewHybridSession(combined, sessionInfo)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := NewHybridSession(combined, sessionInfo)
		if err != nil {
			b.Fatalf("NewHybridSession failed: %v", err)
		}
	}
}

// BenchmarkNewHybridSession_Uncached measures NewHybridSession calls with
// unique sessionInfo to avoid cache hits.
func BenchmarkNewHybridSession_Uncached(b *testing.B) {
	combined := make([]byte, 32)
	rand.Read(combined)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sessionInfo := []byte("bench-session-" + string(byte(i%255)))
		_, err := NewHybridSession(combined, sessionInfo)
		if err != nil {
			b.Fatalf("NewHybridSession failed: %v", err)
		}
	}
}
