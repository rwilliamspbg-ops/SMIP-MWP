// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright 2026 rwilliamspbg-ops
// Security hardening for SMIP-MWP cryptographic operations

package crypto

import (
	"sync/atomic"
	"time"
)

var globalSeqCounter atomic.Uint64

// SecurityConfig holds security parameters for SMIP operations
type SecurityConfig struct {
	MaxReplayWindow  uint64
	HandshakeTimeout time.Duration
	RateLimitPerSec  int
}

var defaultSecurityConfig = SecurityConfig{
	MaxReplayWindow:  MaxReplayWindow,
	HandshakeTimeout: HandshakeTimeout,
	RateLimitPerSec:  10_000_000,
}

// CheckSequenceNumberOverflow detects counter wraparound
func CheckSequenceNumberOverflow(seq uint64, maxSeq uint64) bool {
	if maxSeq == 0 {
		return false
	}
	if seq >= maxSeq {
		return true
	}
	return false
}

// IncrementGlobalSeq increments the global sequence counter
func IncrementGlobalSeq() uint64 {
	return globalSeqCounter.Add(1)
}

// DoSThrottle limits packet processing rate to prevent amplification attacks
type DoSThrottle struct {
	lastPacketTime int64 // Unix timestamp in nanoseconds
	rateLimitNs    int64 // Allowed packets per second (e.g., 1M)
	windowNs       int64 // Sliding window size in nanoseconds
}

func NewDoSThrottle(ratePerSec int) *DoSThrottle {
	return &DoSThrottle{
		rateLimitNs: int64(1_000_000_000 / ratePerSec),
		windowNs:    int64(1e9), // 1 second window
	}
}

// AllowPacket checks if packet processing is allowed under DoS protection
func (d *DoSThrottle) AllowPacket() bool {
	now := time.Now().UnixNano()
	lastSeen := atomic.LoadInt64(&d.lastPacketTime)

	if lastSeen == 0 || now-lastSeen >= d.rateLimitNs {
		atomic.StoreInt64(&d.lastPacketTime, now)
		return true
	}
	return false
}
