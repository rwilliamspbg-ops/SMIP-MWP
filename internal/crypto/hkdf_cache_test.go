package crypto

import (
	"fmt"
	"testing"
)

func TestHKDFCacheBoundsAndDeterminism(t *testing.T) {
	// Swap in a small cache for testing and restore afterwards
	orig := hkdfCache
	hkdfCache = newHKDFCache(1, 16)
	defer func() { hkdfCache = orig }()

	combined := []byte("combined-base")
	for i := 0; i < 32; i++ {
		info := []byte(fmt.Sprintf("session-%d", i))
		if err := PrederiveSession(combined, info); err != nil {
			t.Fatalf("PrederiveSession failed: %v", err)
		}
	}

	if got := hkdfCache.Len(); got > 16 {
		t.Fatalf("hkdfCache exceeded max: got=%d want<=16", got)
	}

	// Determinism: same inputs yield same cache key
	k1 := deriveCacheKey([]byte("a"), []byte("b"))
	k2 := deriveCacheKey([]byte("a"), []byte("b"))
	if k1 != k2 {
		t.Fatalf("deriveCacheKey not deterministic")
	}
}
