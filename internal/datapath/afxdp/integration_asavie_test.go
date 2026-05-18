//go:build withafxdp && asavie
// +build withafxdp,asavie

package afxdp

import (
	"os"
	"testing"
)

// TestAsavieIntegrationSanity is a minimal integration check. It always
// verifies that NewUMEM returns a usable object. The live socket portion
// runs only when RUN_XDP_INTEGRATION=1 and XDP_IFACE is set — this avoids
// requiring privileged hardware in CI.
func TestAsavieIntegrationSanity(t *testing.T) {
	u, err := NewUMEM(128, 2048)
	if err != nil {
		t.Fatalf("NewUMEM failed: %v", err)
	}
	if u == nil {
		t.Fatalf("NewUMEM returned nil UMEM")
	}

	if os.Getenv("RUN_XDP_INTEGRATION") != "1" {
		t.Skip("live XDP integration skipped; set RUN_XDP_INTEGRATION=1 and XDP_IFACE to run")
	}

	iface := os.Getenv("XDP_IFACE")
	if iface == "" {
		t.Skip("XDP_IFACE not set; skipping live socket create")
	}

	s, err := NewXDPSocket(iface, 0, u)
	if err != nil {
		t.Fatalf("NewXDPSocket failed for iface %s: %v", iface, err)
	}
	if s == nil {
		t.Fatalf("NewXDPSocket returned nil socket")
	}
	if err := s.Close(); err != nil {
		t.Fatalf("socket Close failed: %v", err)
	}
}
