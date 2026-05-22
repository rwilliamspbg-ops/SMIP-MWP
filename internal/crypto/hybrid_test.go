package crypto

import (
    "testing"
)

func TestPrederiveAndNewHybridSessionConsistency(t *testing.T) {
    combined := []byte("combined-secret-for-test-0123456789")
    info := []byte("session-info-test")

    // Prederive should succeed and populate cache
    if err := PrederiveSession(combined, info); err != nil {
        t.Fatalf("PrederiveSession failed: %v", err)
    }

    // NewHybridSession should succeed and return a usable session
    s, err := NewHybridSession(combined, info)
    if err != nil {
        t.Fatalf("NewHybridSession failed: %v", err)
    }
    if s == nil {
        t.Fatalf("NewHybridSession returned nil session")
    }

    // second Prederive should be idempotent
    if err := PrederiveSession(combined, info); err != nil {
        t.Fatalf("PrederiveSession (2) failed: %v", err)
    }
}
