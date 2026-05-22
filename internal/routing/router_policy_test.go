package routing

import (
	"testing"
)

func TestComputeFlowKeyDeterministic(t *testing.T) {
	r := NewRouter()
	a := makeID("src-aaa")
	b := makeID("dst-bbb")
	k1 := r.computeFlowKey(a, b, 42)
	k2 := r.computeFlowKey(a, b, 42)
	if k1 != k2 {
		t.Fatalf("computeFlowKey not deterministic: %x != %x", k1, k2)
	}
}

func TestUpdatePolicyAndLookupPriority(t *testing.T) {
	r := NewRouter()
	src := makeID("src-1")
	dst := makeID("dst-1")
	next := makeID("next-1")

	// Ensure default policy exists and has lower priority
	def, err := r.LookupPolicy(src, dst, 0)
	if err != nil {
		t.Fatalf("LookupPolicy default failed: %v", err)
	}
	if def.Priority <= 0 {
		t.Fatalf("default policy priority unexpected: %d", def.Priority)
	}

	// Update policy (should set Priority=1, highest)
	r.UpdatePolicy(src, dst, 7, next, 3)
	p, err := r.LookupPolicy(src, dst, 7)
	if err != nil {
		t.Fatalf("LookupPolicy after update failed: %v", err)
	}
	if p.Priority != 1 {
		t.Fatalf("expected update to set Priority=1, got %d", p.Priority)
	}
	if p.QueueID != 3 {
		t.Fatalf("expected QueueID 3, got %d", p.QueueID)
	}
}
