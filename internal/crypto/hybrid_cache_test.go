// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright 2026 rwilliamspbg-ops
//
// This file is part of SMIP-MWP.
// SMIP-MWP is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation; either version 3 of the License, or (at your option) any later version.
// See the LICENSE file in the project root for details.

package crypto

import "testing"

func TestLRUCacheBoundsAndEvictsLeastRecent(t *testing.T) {
	cache := newLRUCache(2)

	var key1 [32]byte
	var key2 [32]byte
	var key3 [32]byte
	key1[0] = 1
	key2[0] = 2
	key3[0] = 3

	cache.Put(key1, hkdfCacheEntry{seqMask: 1})
	cache.Put(key2, hkdfCacheEntry{seqMask: 2})

	if _, ok := cache.Get(key1); !ok {
		t.Fatal("expected key1 to be cached")
	}

	cache.Put(key3, hkdfCacheEntry{seqMask: 3})

	if got := cache.Len(); got != 2 {
		t.Fatalf("expected cache size 2, got %d", got)
	}
	if _, ok := cache.Get(key2); ok {
		t.Fatal("expected key2 to be evicted")
	}
	if _, ok := cache.Get(key1); !ok {
		t.Fatal("expected key1 to remain cached")
	}
	if _, ok := cache.Get(key3); !ok {
		t.Fatal("expected key3 to be cached")
	}
}
