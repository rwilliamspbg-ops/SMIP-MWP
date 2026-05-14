# SMIP-MWP Week 1 Quick Start Guide

## Objective: Complete Hybrid Handshake

**Timeline**: May 14-20 (This week!)

**Success Criteria**:
- ✅ ML-KEM integrated (real or interim placeholder)
- ✅ Handshake state machine fully implemented
- ✅ crypto/kex_test.go: Full handshake flow passes
- ✅ Handshake completes in <50ms on LAN

---

## Task 1: ML-KEM Integration (May 14-15)

### What Needs to Happen

Replace the ML-KEM stub (random bytes) in [crypto/kex.go](crypto/kex.go) with real key generation.

### Option A: Real ML-KEM (Go 1.24+)

**Check Go Version**:
```bash
go version  # Must be >= 1.24.0
```

**If Go 1.24+ is available**:

1. Add import in `crypto/kex.go`:
```go
import (
    "crypto/mlkem"  // NEW
    // ... other imports
)
```

2. Update `NewHybridKEX()` function (lines ~50-70):
```go
// OLD CODE (REMOVE):
// h.mlkemPub = make([]byte, 1184)
// h.mlkemPriv = make([]byte, 2400)
// if _, err := io.ReadFull(rng, h.mlkemPub); err != nil { ... }

// NEW CODE (ADD):
mlkemKey, err := mlkem.GenerateKey768(rng)
if err != nil {
    return nil, fmt.Errorf("kex: mlkem key gen: %w", err)
}
h.mlkemPub = mlkemKey.PublicKey().Bytes()
h.mlkemPriv = mlkemKey.Bytes()
```

3. Test compilation:
```bash
cd /workspaces/SMIP-MWP
go mod tidy
go build ./...
```

### Option B: Interim with cloudflare/circl (Go <1.24)

If Go 1.24 is not available:

1. Add to `go.mod`:
```
require github.com/cloudflare/circl v1.5.0
```

2. Update `crypto/kex.go`:
```go
import (
    "github.com/cloudflare/circl/kem/mlkem/mlkem768"
)

// In NewHybridKEX():
seed := make([]byte, 64)
if _, err := io.ReadFull(rng, seed); err != nil {
    return nil, err
}
ek, dk, err := mlkem768.Keygen(seed)
if err != nil {
    return nil, fmt.Errorf("kex: mlkem key gen: %w", err)
}
h.mlkemPub = ek
h.mlkemPriv = dk
```

3. Test:
```bash
go mod tidy
go build ./...
```

**✅ Task 1 Complete When**: Code compiles and `go test crypto/...` passes (existing tests).

---

## Task 2: Handshake State Machine (May 16-18)

### Current State

File: `internal/crypto/hybrid.go` has skeleton:
```go
type HybridSession struct {
    State      SessionState
    Aead       cipher.AEAD
    // ... incomplete
}

func (s *HybridSession) ProcessHandshakeMessage(...) error {
    // ... partial implementation
}
```

### What Needs to Happen

**Complete the state machine** with all transitions + timeout + replay protection.

### Step 1: Add Timeout Tracking

```go
type HybridSession struct {
    State          SessionState
    Aead           cipher.AEAD
    NonceBase      [NonceSize]byte
    SeqMask        uint64
    
    // NEW: Handshake timeout + retry
    HandshakeStart time.Time
    RetryCount     int
    SeqWindow      uint64  // For replay detection
}

const (
    HandshakeTimeout = 30 * time.Second
    MaxRetries       = 3
)
```

### Step 2: Complete State Transitions

In `ProcessHandshakeMessage()`:

```go
func (s *HybridSession) ProcessHandshakeMessage(peerXPublic, peerMLPublic []byte) error {
    // Check timeout
    if !s.HandshakeStart.IsZero() && time.Since(s.HandshakeStart) > HandshakeTimeout {
        s.State = UNINITIALIZED  // Reset
        return fmt.Errorf("handshake timeout")
    }

    switch s.State {
    case UNINITIALIZED:
        return fmt.Errorf("session not initialized; call InitiateHandshake first")

    case AWAITING_PEER_PUBKEY:
        if len(peerXPublic) == 0 || len(peerMLPublic) == 0 {
            s.RetryCount++
            if s.RetryCount >= MaxRetries {
                s.State = UNINITIALIZED
                return fmt.Errorf("max retries exceeded")
            }
            return fmt.Errorf("missing peer public keys")
        }

        // Perform hybrid key exchange
        sharedSecret, err := s.performHybridKeyExchange(peerXPublic, peerMLPublic)
        if err != nil {
            return fmt.Errorf("key exchange failed: %w", err)
        }

        // Derive session keys
        sessionInfo := append([]byte("MWP_SESSION"), s.LocalXPrivate...)
        newSession, err := NewHybridSessionFromMaterial(sharedSecret, sessionInfo)
        if err != nil {
            return fmt.Errorf("session key derivation failed: %w", err)
        }

        // Move to ESTABLISHED
        s.Aead = newSession.Aead
        s.NonceBase = newSession.NonceBase
        s.SeqMask = newSession.SeqMask
        s.State = ESTABLISHED
        return nil

    case ESTABLISHED:
        return fmt.Errorf("session already established")

    default:
        return fmt.Errorf("unknown state: %v", s.State)
    }
}
```

### Step 3: Add Replay Protection

```go
func (s *HybridSession) CheckReplay(seq uint64) bool {
    // Simple: check if seq bit already set in 64-bit window
    mask := uint64(1) << (seq % 64)
    if s.SeqWindow&mask != 0 {
        return true  // Replay detected
    }
    s.SeqWindow |= mask
    return false
}
```

### Step 4: Add Helper Functions

```go
func (s *HybridSession) InitiateHandshake(rand io.Reader) error {
    if s.State != UNINITIALIZED {
        return fmt.Errorf("handshake already initiated")
    }
    
    s.LocalXPrivate = make([]byte, 32)
    s.LocalMLPrivate = make([]byte, 2400)  // ML-KEM-768 private key size
    
    if _, err := io.ReadFull(rand, s.LocalXPrivate); err != nil {
        return err
    }
    if _, err := io.ReadFull(rand, s.LocalMLPrivate); err != nil {
        return err
    }
    
    s.HandshakeStart = time.Now()
    s.State = AWAITING_PEER_PUBKEY
    return nil
}

func (s *HybridSession) performHybridKeyExchange(peerX, peerML []byte) ([]byte, error) {
    // X25519
    var peerX25519 [32]byte
    copy(peerX25519[:], peerX)
    var x25519SS [32]byte
    curve25519.ScalarMult(&x25519SS, (*[32]byte)(s.LocalXPrivate), &peerX25519)
    
    // ML-KEM (stub for now; will be real with crypto/mlkem)
    mlkemSS := make([]byte, 32)
    copy(mlkemSS, peerML[:32])  // Placeholder
    
    // Combine
    combined := append(x25519SS[:], mlkemSS...)
    return combined, nil
}
```

**✅ Task 2 Complete When**: 
- `internal/crypto/hybrid.go` compiles without errors
- State transitions follow: UNINITIALIZED → AWAITING_PEER_PUBKEY → ESTABLISHED
- Timeout enforced after 30s
- Replay detection functional

---

## Task 3: Unit Tests (May 19-20)

### Create New Test File

**New File**: `crypto/kex_test.go`

```go
package crypto

import (
    "bytes"
    "crypto/rand"
    "testing"
)

func TestHybridHandshakeFullFlow(t *testing.T) {
    // 1. Alice generates keypair
    alice, err := NewHybridKEX(rand.Reader)
    if err != nil {
        t.Fatalf("Alice key gen failed: %v", err)
    }
    alicePub := alice.PublicKey()

    // 2. Bob generates keypair
    bob, err := NewHybridKEX(rand.Reader)
    if err != nil {
        t.Fatalf("Bob key gen failed: %v", err)
    }
    bobPub := bob.PublicKey()

    // 3. Alice performs handshake
    aliceSecret, err := alice.Handshake(bobPub)
    if err != nil {
        t.Fatalf("Alice handshake failed: %v", err)
    }

    // 4. Bob performs handshake
    bobSecret, err := bob.Handshake(alicePub)
    if err != nil {
        t.Fatalf("Bob handshake failed: %v", err)
    }

    // 5. Verify shared secrets match (CRITICAL!)
    if !bytes.Equal(aliceSecret, bobSecret) {
        t.Fatalf("Shared secrets don't match:\n  Alice: %x\n  Bob:   %x", aliceSecret, bobSecret)
    }

    // 6. Verify keys are long enough for session
    if len(aliceSecret) < 64 {
        t.Fatalf("Shared secret too short: %d bytes", len(aliceSecret))
    }

    t.Logf("✅ Full handshake succeeded")
}

func TestStateMachineEnforcement(t *testing.T) {
    sess := &HybridSession{State: UNINITIALIZED}

    // Try to process handshake before InitiateHandshake
    err := sess.ProcessHandshakeMessage([]byte{}, []byte{})
    if err == nil {
        t.Fatalf("Should reject handshake in UNINITIALIZED state")
    }

    // Initiate
    if err := sess.InitiateHandshake(rand.Reader); err != nil {
        t.Fatalf("InitiateHandshake failed: %v", err)
    }
    if sess.State != AWAITING_PEER_PUBKEY {
        t.Fatalf("Expected AWAITING_PEER_PUBKEY, got %v", sess.State)
    }

    t.Logf("✅ State machine enforcement working")
}

func TestReplayProtection(t *testing.T) {
    sess := &HybridSession{}

    // First packet: OK
    if sess.CheckReplay(1) {
        t.Fatalf("First packet should not be flagged as replay")
    }

    // Duplicate: Replay detected
    if !sess.CheckReplay(1) {
        t.Fatalf("Duplicate packet should be flagged as replay")
    }

    t.Logf("✅ Replay protection working")
}

func TestHandshakeTimeout(t *testing.T) {
    sess := &HybridSession{}
    sess.InitiateHandshake(rand.Reader)
    
    // Artificially set timeout
    sess.HandshakeStart = time.Now().Add(-31 * time.Second)
    
    err := sess.ProcessHandshakeMessage([]byte{1}, []byte{1})
    if err == nil || !strings.Contains(err.Error(), "timeout") {
        t.Fatalf("Should detect handshake timeout")
    }

    t.Logf("✅ Handshake timeout working")
}
```

### Run Tests

```bash
cd /workspaces/SMIP-MWP
go test -v ./crypto/...
```

**Expected Output**:
```
=== RUN   TestHybridHandshakeFullFlow
    kex_test.go:XX: ✅ Full handshake succeeded
--- PASS: TestHybridHandshakeFullFlow (0.015s)
=== RUN   TestStateMachineEnforcement
    kex_test.go:XX: ✅ State machine enforcement working
--- PASS: TestStateMachineEnforcement (0.002s)
=== RUN   TestReplayProtection
    kex_test.go:XX: ✅ Replay protection working
--- PASS: TestReplayProtection (0.001s)
=== RUN   TestHandshakeTimeout
    kex_test.go:XX: ✅ Handshake timeout working
--- PASS: TestHandshakeTimeout (0.003s)

ok      github.com/rwilliamspbg-ops/smip-mwp-forge/crypto  0.021s
```

**✅ Task 3 Complete When**: All tests pass + coverage >90%

---

## Validation Checklist (End of Week 1)

- [ ] Go version ≥1.24 OR circl interim ML-KEM working
- [ ] `crypto/kex.go` compiles
- [ ] `internal/crypto/hybrid.go` compiles
- [ ] All state machine tests pass
- [ ] Replay protection verified
- [ ] Timeout enforcement verified
- [ ] Handshake latency <50ms (measure with BenchmarkHybridHandshake)
- [ ] KEX produces matching shared secrets
- [ ] No panics or segfaults

---

## Performance Benchmark (Optional but Encouraged)

Add to `crypto/kex_test.go`:

```go
func BenchmarkHybridHandshake(b *testing.B) {
    alice, _ := NewHybridKEX(rand.Reader)
    alicePub := alice.PublicKey()
    bob, _ := NewHybridKEX(rand.Reader)
    bobPub := bob.PublicKey()

    b.ReportAllocs()
    b.ResetTimer()

    for i := 0; i < b.N; i++ {
        alice.Handshake(bobPub)
    }
}
```

Run:
```bash
go test -bench=BenchmarkHybridHandshake -benchmem ./crypto/...
```

**Expected**: ~10-30ms per handshake (will improve with real ML-KEM hardware acceleration)

---

## Daily Progress (May 14-20)

### Monday, May 14
**Goal**: ML-KEM integration complete

- [ ] Check Go version + decide strategy (real vs. interim)
- [ ] Update `crypto/kex.go` with real/interim ML-KEM
- [ ] `go build ./...` passes
- [ ] Basic test: generate keys + verify sizes (1184B pub, 2400B priv)

**EOD Status**: ✅ ML-KEM working

### Tuesday, May 15
**Goal**: Handshake flow skeleton

- [ ] Implement `InitiateHandshake()`
- [ ] Implement `ProcessHandshakeMessage()` state transitions
- [ ] Code compiles (may have stub test failures)

**EOD Status**: ✅ State machine structure

### Wednesday, May 16
**Goal**: Timeout + replay protection

- [ ] Add timeout tracking + abort logic
- [ ] Implement `CheckReplay()` function
- [ ] Add `EncryptInPlace()` integration (call from session)

**EOD Status**: ✅ Timeout + replay working

### Thursday, May 17
**Goal**: Start unit tests

- [ ] Create `crypto/kex_test.go`
- [ ] Write `TestHybridHandshakeFullFlow()`
- [ ] Measure handshake latency

**EOD Status**: ✅ Full handshake test passing

### Friday, May 18-20
**Goal**: Complete test suite + validation

- [ ] Add state machine tests
- [ ] Add timeout tests
- [ ] Add replay tests
- [ ] Coverage >90%
- [ ] All tests passing + <50ms handshake

**EOD Status**: ✅ Phase 1 Handshake COMPLETE

---

## Troubleshooting Guide

### Problem: Compile Error - "crypto/mlkem not found"

**Solution**: Go <1.24 detected.
```bash
# Use interim circl version
go get github.com/cloudflare/circl@v1.5.0
# Update crypto/kex.go to use circl (see Option B above)
```

### Problem: "Shared secrets don't match in test"

**Possible Causes**:
1. X25519 key exchange wrong (check scalar mult logic)
2. ML-KEM stub produces different random each time (expected; will fix with real crypto)
3. HKDF combiner not deterministic

**Debug**:
```go
// Print intermediate values
t.Logf("Alice x25519 secret: %x", aliceSecret[:32])
t.Logf("Bob x25519 secret:   %x", bobSecret[:32])
```

### Problem: State machine test fails

**Check**:
- Is `State` properly transitioning? (print state at each step)
- Is timeout calculated correctly? (check `time.Since()`)
- Is `RetryCount` incrementing?

### Problem: Handshake takes >50ms

**Profile**:
```bash
go test -cpuprofile=cpu.prof -bench=BenchmarkHybridHandshake ./crypto/...
go tool pprof cpu.prof  # Analyze bottleneck
```

**Likely culprits**: 
- ML-KEM key generation (expected to be slow until hardware acceleration)
- HKDF expansion (should be fast; check if loop is too large)

---

## Success Signoff

When all tests pass and validation checklist complete, Week 1 is **DONE**. 

**Sign-off**:
```bash
# Final validation
go test -v ./crypto/... 
go build ./cmd/...  # Should compile (even if not fully implemented)

# Report
echo "✅ Week 1 Complete: Hybrid Handshake Fully Implemented"
```

---

## Next Steps (Week 2 Preview)

Once Week 1 complete:
- [ ] Create `cmd/mohawk-node/main.go` entry point
- [ ] Implement UDP overlay transport (`internal/transport/udp.go`)
- [ ] Integrate with handshake (full flow: receive packet → initiate handshake → establish session → forward)
- [ ] Measure end-to-end latency + throughput

**Week 2 Goal**: UDP forwarding at 1k+ pps with <5ms latency.

Good luck! 🚀
