package crypto

import (
	"crypto/rand"
	"testing"
)

func TestPrederiveSessionPopulatesCache(t *testing.T) {
	combined := make([]byte, 32)
	if _, err := rand.Read(combined); err != nil {
		t.Fatalf("rand: %v", err)
	}
	sessionInfo := []byte("prederive-test-session")

	before := hkdfCache.Len()
	if err := PrederiveSession(combined, sessionInfo); err != nil {
		t.Fatalf("PrederiveSession: %v", err)
	}
	after := hkdfCache.Len()
	if after <= before {
		t.Fatalf("expected hkdfCache.Len() to increase, before=%d after=%d", before, after)
	}
}
