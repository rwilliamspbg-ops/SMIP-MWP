package crypto

import (
	"crypto/sha256"
	"testing"
)

func TestHybridSessionRoundTrip(t *testing.T) {
	secret := make([]byte, 64)
	sessionInfo := []byte("integration-test-flow")

	combined := sha256.Sum256(append(secret, sessionInfo...))
	sess, err := NewHybridSession(combined[:], sessionInfo)
	if err != nil {
		t.Fatalf("NewHybridSession: %v", err)
	}

	original := make([]byte, 128)
	copy(original, []byte("test packet data here"))

	encrypted, err := sess.Encrypt(original, 0)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	decrypted, err := sess.Decrypt(encrypted, 0)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}

	want := "test packet data here"
	if len(decrypted) < len(want) || string(decrypted[:len(want)]) != want {
		t.Errorf("Decrypted data mismatch: got=%q wantPrefix=%q", string(decrypted), want)
	}
}
