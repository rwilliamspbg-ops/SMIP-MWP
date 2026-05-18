// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright 2026 rwilliamspbg-ops
//
// This file is part of SMIP-MWP.
// SMIP-MWP is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation; either version 3 of the License, or (at your option) any later version.
// See the LICENSE file in the project root for details.

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
