package crypto

import (
	"testing"
    "time"
)

func TestHybridKEXStateCleanupAndSeqWindow(t *testing.T) {
    var sid [16]byte
    s := NewHybridKEXState(sid)
    if s == nil {
        t.Fatalf("NewHybridKEXState returned nil")
    }

    // simulate timeout by setting the deadline in the past; ensure timeout is in the past
    s.timeout = time.Now().Add(-1 * time.Second)
    if !s.timeout.Before(time.Now()) {
        t.Fatalf("timeout not in the past")
    }

    // Cleanup should zero time fields
    s.Cleanup()
    if !s.kexStarted.IsZero() || !s.timeout.IsZero() {
        t.Fatalf("Cleanup did not reset times")
    }

    // Seq window: increment many times and ensure no false positives
    for i := 0; i < MaxReplayWindow+5; i++ {
        if _, err := s.IncrementSeqCounter(); err != nil {
            t.Fatalf("IncrementSeqCounter failed at i=%d: %v", i, err)
        }
    }
}
